package helpers

import (
	"context"
	"encoding/json"
	"errors"
	clolog "github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"strconv"
	"time"
)

const maxLogsLimit = 5

type logExplorationAPI struct {
	deployment *apps.Deployment
	service    *corev1.Service
	tc         *E2ETestFramework
}

func (l *logExplorationAPI) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	panic("implement me")
}
func (l *logExplorationAPI) HasApplicationLogs(timeToWait time.Duration) (bool, error) {

	logs, err := l.getLogs()
	if err != nil {
		return false, err
	}
	if len(logs) <= 1 {
		time.Sleep(180 * time.Second)
		logs, err = l.getLogs()
		if err == nil && len(logs) > 1 {
			return true, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (l *logExplorationAPI) getLogs() ([]interface{}, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=elasticsearch",
	}

	pods, err := l.tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("No pods found for elasticsearch")
	}
	clolog.V(3).Info("Pod ", "PodName", pods.Items[0].Name)
	stdout, err := l.tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "elasticsearch", []string{"curl", "http://logexplorationapi:8080/logs?maxlogs=" + strconv.Itoa(maxLogsLimit)})
	if err != nil {
		log.Println("error while fetching logs is: ", err)
		return nil, err
	}
	var jsonObjs map[string]interface{}

	err = json.Unmarshal([]byte(stdout), &jsonObjs)
	if err != nil {
		log.Println("cannot unmarshall the response body", err)
		return nil, err
	}
	objSlice, ok := jsonObjs["Logs"].([]interface{})
	if !ok {
		log.Println("cannot convert the JSON objects", err)
		return nil, err
	}
	return objSlice, nil
}
func (l *logExplorationAPI) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	panic("implement me")
}

func (l *logExplorationAPI) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	panic("implement me")
}

func (l *logExplorationAPI) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	panic("implement me")
}

func (l *logExplorationAPI) RetrieveLogs() (map[string]string, error) {
	panic("implement me")
}

func (l *logExplorationAPI) ClusterLocalEndpoint() string {
	panic("implement me")
}

func (tc *E2ETestFramework) DeleteLogExplorationAPI(api logExplorationAPI) error {
	var zerograce int64
	opts := metav1.DeleteOptions{
		GracePeriodSeconds: &zerograce,
	}
	return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(context.TODO(), api.deployment.Name, opts)
}

func (tc *E2ETestFramework) DeployLogExplorationAPI() (deployment *apps.Deployment, err error) {
	logExploration := &logExplorationAPI{
		tc: tc,
	}
	httpGetStructLivenessProbe := &corev1.HTTPGetAction{
		Path: "/ready",
		Port: intstr.FromInt(8080),
	}
	httpGetStructReadinessProbe := &corev1.HTTPGetAction{
		Path: "/health",
		Port: intstr.FromInt(8080),
	}
	handlerStructLivenessProbe := &corev1.Handler{
		HTTPGet: httpGetStructLivenessProbe,
	}
	handlerStructReadinessProbe := &corev1.Handler{
		HTTPGet: httpGetStructReadinessProbe,
	}
	container := corev1.Container{
		Name:            "logexplorationapi",
		Image:           "quay.io/openshift-logging/log-exploration-api:latest",
		ImagePullPolicy: corev1.PullAlways,
		Env: []corev1.EnvVar{
			{Name: "ES_ADDR", Value: "https://elasticsearch.openshift-logging:9200"},
			{Name: "ES_CERT", Value: "/etc/openshift/elasticsearch/secret/tls.crt"},
			{Name: "ES_KEY", Value: "/etc/openshift/elasticsearch/secret/tls.key"},
			{Name: "ES_TLS", Value: "true"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "certificates", MountPath: "/etc/openshift/elasticsearch/secret"},
		},

		LivenessProbe: &corev1.Probe{
			Handler:             *handlerStructLivenessProbe,
			InitialDelaySeconds: 10,
			PeriodSeconds:       3,
			FailureThreshold:    30,
		},
		ReadinessProbe: &corev1.Probe{
			Handler:             *handlerStructReadinessProbe,
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
		Volumes: []corev1.Volume{
			{Name: "certificates", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "fluentd", DefaultMode: utils.GetInt32(420)}}},
		},
	}
	dOpts := metav1.CreateOptions{}
	logExplorationApiDeployment := k8shandler.NewDeployment(
		container.Name,
		OpenshiftLoggingNS,
		container.Name,
		container.Name,
		podSpec,
	)
	logExplorationApiDeployment, err = tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Create(context.TODO(), logExplorationApiDeployment, dOpts)

	if err != nil {
		return nil, err
	}
	service := factory.NewService(
		container.Name,
		OpenshiftLoggingNS,
		container.Name,
		[]corev1.ServicePort{
			{
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
	tc.AddCleanup(func() error {
		_, err := oc.Exec().
			WithNamespace(OpenshiftLoggingNS).
			WithPodGetter(oc.Get().
				WithNamespace(OpenshiftLoggingNS).
				Pod().
				OutputJsonpath("{.items[0].metadata.name}")).
			Container(container.Name).
			Run()
		return err
	})
	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(context.TODO(), logExplorationApiDeployment.Name, opts)
	})
	sOpts := metav1.CreateOptions{}
	service, err = tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Create(context.TODO(), service, sOpts)
	if err != nil {
		return nil, err
	}

	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Delete(context.TODO(), service.Name, opts)
	})
	logExploration.deployment = logExplorationApiDeployment
	logExploration.service = service
	name := logExplorationApiDeployment.GetName()
	tc.LogStores[name] = logExploration
	return logExplorationApiDeployment, tc.waitForDeployment(OpenshiftLoggingNS, logExplorationApiDeployment.Name, defaultRetryInterval, defaultTimeout)
}

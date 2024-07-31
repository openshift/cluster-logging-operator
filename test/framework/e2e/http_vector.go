package e2e

import (
	"context"
	"errors"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"path"
	"strings"
	"time"
)

const (
	HttpReceiver   = "http-receiver"
	vectorHttpConf = "" +
		`[sources.my_source]
type = "http_server"
address = "0.0.0.0:8090"
decoding.codec = "json"
framing.method = "newline_delimited"

[transforms.route_logs]
type = "route"
inputs = ["my_source"]
route.audit = '.log_type == "audit"'
route.container = 'exists(.kubernetes)'
route.journal = '!exists(.kubernetes)'

[sinks.container]
inputs = ["route_logs.container"]
type = "file"
path = "/tmp/container/{{kubernetes.namespace_name}}_{{kubernetes.pod_name}}_{{kubernetes.container_name}}.json"
		
[sinks.container.encoding]
codec = "json"

[sinks.out_journal]
inputs = ["route_logs.journal"]
type = "file"
path = "/tmp/journal/journal.json"
		
[sinks.out_journal.encoding]
codec = "json"

[sinks.out_audit]
inputs = ["route_logs.audit"]
type = "file"
path = "/tmp/audit/audit.json"
		
[sinks.out_audit.encoding]
codec = "json"
`
)

type VectorHttpReceiverLogStore struct {
	*apps.Deployment
	tc *E2ETestFramework
}

func (tc *E2ETestFramework) DeployHttpReceiver(ns string) (deployment *VectorHttpReceiverLogStore, err error) {
	logStore := &VectorHttpReceiverLogStore{
		tc: tc,
	}
	serviceAccount, err := tc.createServiceAccount(ns, HttpReceiver)
	if err != nil {
		log.Error(err, "Unable to create service account")
		return nil, err
	}
	container := corev1.Container{
		Name:  HttpReceiver,
		Image: utils.GetComponentImage(constants.VectorName),
		Ports: []corev1.ContainerPort{
			{Name: "http", ContainerPort: 8090},
		},
		Command:         []string{"bash", path.Join("/opt", vector.RunVectorFile)},
		ImagePullPolicy: corev1.PullAlways,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.GetPtr(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			RunAsNonRoot: utils.GetPtr(true),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				ReadOnly:  true,
				MountPath: "/etc/vector",
			},
			{
				Name:      "config",
				ReadOnly:  true,
				MountPath: "/opt",
			},
		},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
		Volumes: []corev1.Volume{
			{
				Name: "config", VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: container.Name,
						},
					},
				},
			},
		},
		ServiceAccountName: serviceAccount.Name,
	}
	opts := metav1.CreateOptions{}
	config := runtime.NewConfigMap(ns, container.Name, map[string]string{
		vector.ConfigFile:    vectorHttpConf,
		vector.RunVectorFile: fmt.Sprintf(vector.RunVectorScript, path.Join("/tmp/vector", ns, container.Name)),
	})
	config, err = tc.KubeClient.CoreV1().ConfigMaps(ns).Create(context.TODO(), config, opts)
	if err != nil {
		log.Error(err, "Unable to create configmap", "configmap.meta", config.ObjectMeta)
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().ConfigMaps(ns).Delete(context.TODO(), config.Name, opts)
	})

	dOpts := metav1.CreateOptions{}
	logStore.Deployment = k8shandler.NewDeployment(
		container.Name,
		ns,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)
	// Add instance label to pod spec template. Service now selects using instance name as well
	logStore.Deployment.Spec.Template.Labels[constants.LabelK8sInstance] = HttpReceiver
	logStore.Deployment.Spec.Template.Labels["vector.dev/exclude"] = "true"

	log.V(1).Info("Deploying http receiver deployment", "namespace", ns, "name", logStore.Deployment.Name)
	logStore.Deployment, err = tc.KubeClient.AppsV1().Deployments(ns).Create(context.TODO(), logStore.Deployment, dOpts)
	if err != nil {
		log.Error(err, "Unable to create Deployment", "meta", logStore.Deployment.ObjectMeta)
		return nil, err
	}

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().Deployments(ns).Delete(context.TODO(), logStore.Deployment.Name, opts)
	})

	service := factory.NewService(
		serviceAccount.Name,
		ns,
		serviceAccount.Name,
		serviceAccount.Name,
		[]corev1.ServicePort{
			{
				Port: 8090,
			},
		},
	)

	sOpts := metav1.CreateOptions{}
	service, err = tc.KubeClient.CoreV1().Services(ns).Create(context.TODO(), service, sOpts)
	if err != nil {
		log.Error(err, "Unable to create service", "meta", service.ObjectMeta)
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Services(ns).Delete(context.TODO(), service.Name, opts)
	})
	tc.LogStores[logStore.Deployment.Name] = logStore
	return logStore, tc.waitForDeployment(ns, logStore.Deployment.Name, defaultRetryInterval, 1*time.Minute)
}

type ContainerLogSimpleMeta struct {
	Namespace     string
	PodName       string
	ContainerName string
}

func NewLogSimpleMeta(parts []string) *ContainerLogSimpleMeta {
	meta := &ContainerLogSimpleMeta{}
	if len(parts) > 0 {
		meta.Namespace = parts[0]
	}
	if len(parts) > 1 {
		meta.PodName = parts[1]
	}
	if len(parts) > 2 {
		meta.ContainerName = strings.TrimSuffix(parts[2], ".json")
	}
	return meta
}

type Query struct {
	Meta  []ContainerLogSimpleMeta
	files []string
}

func (v VectorHttpReceiverLogStore) ListNamespaces() (namespaces []string) {
	q, err := v.Query(nil)
	if err != nil {
		log.Error(err, "Error checking receiver")
	}
	for _, m := range q.Meta {
		namespaces = append(namespaces, m.Namespace)
	}
	return namespaces
}

func (v VectorHttpReceiverLogStore) ListContainers() (containers []string) {
	q, err := v.Query(nil)
	if err != nil {
		log.Error(err, "Error checking receiver")
	}
	for _, m := range q.Meta {
		containers = append(containers, m.ContainerName)
	}
	return containers
}

func isFileDoesNotExistError(out string) bool {
	return strings.Contains(out, "No such file or directory")
}

func (v VectorHttpReceiverLogStore) ListJournalLogs() ([]types.JournalLog, error) {
	result, err := v.RunCmd("head -n 10 /tmp/journal/journal.json", nil)
	if err != nil {
		return nil, err
	}
	out := "[" + strings.TrimRight(strings.Replace(result, "\n", ",", -1), ",") + "]"
	return types.ParseJournalLogs[types.JournalLog](out)
}

func (v VectorHttpReceiverLogStore) RunCmd(cmd string, timeout *time.Duration) (string, error) {
	if timeout == nil {
		timeout = utils.GetPtr(2 * time.Minute)
	}
	options := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", constants.LabelK8sInstance, HttpReceiver),
	}
	pods, err := v.tc.KubeClient.CoreV1().Pods(v.Namespace).List(context.TODO(), options)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", errors.New("No pods found for receiver")
	}
	result := ""
	err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, *timeout, true, func(cxt context.Context) (done bool, err error) {
		if result, err = v.tc.PodExec(v.Namespace, pods.Items[0].Name, "", strings.Split(cmd, " ")); err != nil {
			log.V(4).Error(err, "Failed to fetch logs from receiver", "name", pods.Items[0].Name, "out", result)
			return false, nil
		}
		return true, nil
	})
	if wait.Interrupted(err) {
		if isFileDoesNotExistError(result) {
			return "", nil
		}
		return "", err
	}
	log.V(3).Info("Raw from query receiver", "response", result)
	return result, nil
}

// Query queries the receiver with a timeout for the request and returns a simple list
// of the meta available
func (v VectorHttpReceiverLogStore) Query(timeout *time.Duration) (*Query, error) {
	q := Query{}
	result, err := v.RunCmd("ls /tmp/container", timeout)
	if err != nil {
		return nil, err
	}
	if result == "" {
		return &q, nil
	}
	q.files = strings.Split(result, "\n")
	log.V(4).Info("Split raw", "files", q.files)
	for _, ns := range q.files {
		if path.Ext(ns) == ".json" {
			parts := strings.Split(strings.TrimPrefix(ns, "/tmp/container/"), "_")
			if len(parts) > 0 {
				q.Meta = append(q.Meta, *NewLogSimpleMeta(parts))
			}
		}
	}
	return &q, nil
}

func (v VectorHttpReceiverLogStore) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	result, err := v.Query(&timeToWait)
	if err != nil {
		return nil, err
	}
	lines := []string{}
	for _, file := range result.files {
		out, err := v.RunCmd(fmt.Sprintf("cat /tmp/container/%s", file), &timeToWait)
		if err != nil {
			return nil, err
		}
		stream := strings.Split(out, "\n")
		for _, line := range stream {
			if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
				lines = append(lines, line)
			} else {
				log.Info("Dropped incomplete JSON line", "line", line)
			}
		}
	}
	return types.ParseLogs(fmt.Sprintf("[%s]", strings.Join(lines, ",")))
}

func (v VectorHttpReceiverLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	result, err := v.Query(&timeToWait)
	if err != nil {
		return false, err
	}
	return len(result.Meta) > 0, nil
}

func (v VectorHttpReceiverLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (v VectorHttpReceiverLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (v VectorHttpReceiverLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (v VectorHttpReceiverLogStore) RetrieveLogs() (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (v VectorHttpReceiverLogStore) ClusterLocalEndpoint() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:8090", v.Deployment.Name, v.Deployment.Namespace)
}

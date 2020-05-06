package helpers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	k8shandler "github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	e2eutil "github.com/openshift/cluster-logging-operator/test/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

const (
	clusterLoggingURI      = "apis/logging.openshift.io/v1/namespaces/openshift-logging/clusterloggings"
	clusterlogforwarderURI = "apis/logging.openshift.io/v1/namespaces/openshift-logging/clusterlogforwarders"
	DefaultCleanUpTimeout  = 60.0 * 2
)

var (
	defaultRetryInterval      time.Duration
	defaultTimeout            time.Duration
	DefaultWaitForLogsTimeout time.Duration
	err                       error
)

func init() {
	if defaultRetryInterval, err = time.ParseDuration("1s"); err != nil {
		panic(err)
	}
	if defaultTimeout, err = time.ParseDuration("5m"); err != nil {
		panic(err)
	}
	if DefaultWaitForLogsTimeout, err = time.ParseDuration("5m"); err != nil {
		panic(err)
	}
}

type LogStore interface {
	//ApplicationLogs returns app logs for a given log store
	ApplicationLogs(timeToWait time.Duration) (string, error)

	HasApplicationLogs(timeToWait time.Duration) (bool, error)

	HasInfraStructureLogs(timeToWait time.Duration) (bool, error)

	HasAuditLogs(timeToWait time.Duration) (bool, error)

	GrepLogs(expr string, timeToWait time.Duration) (string, error)
}

type E2ETestFramework struct {
	RestConfig     *rest.Config
	KubeClient     *kubernetes.Clientset
	ClusterLogging *cl.ClusterLogging
	CleanupFns     []func() error
	LogStore       LogStore
}

func NewE2ETestFramework() *E2ETestFramework {
	client, config := newKubeClient()
	framework := &E2ETestFramework{
		RestConfig: config,
		KubeClient: client,
	}
	return framework
}

func (tc *E2ETestFramework) AddCleanup(fn func() error) {
	tc.CleanupFns = append(tc.CleanupFns, fn)
}

func (tc *E2ETestFramework) DeployLogGenerator() error {
	namespace := tc.CreateTestNamespace()
	container := corev1.Container{
		Name:            "log-generator",
		Image:           "busybox",
		ImagePullPolicy: corev1.PullAlways,
		Args:            []string{"sh", "-c", "i=0; while true; do echo $i: My life is my message; i=$((i+1)) ; sleep 1; done"},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
	}
	deployment := k8shandler.NewDeployment("log-generator", namespace, "log-generator", "test", podSpec)
	logger.Infof("Deploying %q to namespace: %q...", deployment.Name, deployment.Namespace)
	deployment, err := tc.KubeClient.Apps().Deployments(namespace).Create(deployment)
	if err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Apps().Deployments(namespace).Delete(deployment.Name, nil)
	})
	return tc.waitForDeployment(namespace, "log-generator", defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) CreateTestNamespace() string {
	name := fmt.Sprintf("clo-test-%d", rand.Intn(10000))
	if value, found := os.LookupEnv("GENERATOR_NS"); found {
		name = value
	}
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := tc.KubeClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		logger.Error(err)
	}
	return name
}

func (tc *E2ETestFramework) WaitFor(component LogComponentType) error {
	switch component {
	case ComponentTypeVisualization:
		return tc.waitForDeployment(OpenshiftLoggingNS, "kibana", defaultRetryInterval, defaultTimeout)
	case ComponentTypeCollector:
		logger.Debugf("Waiting for %v", component)
		return e2eutil.WaitForDaemonSet(&testing.T{}, tc.KubeClient, OpenshiftLoggingNS, "fluentd", defaultRetryInterval, defaultTimeout)
	case ComponentTypeStore:
		return tc.waitForElasticsearchPods(defaultRetryInterval, defaultTimeout)
	}
	return fmt.Errorf("Unable to waitfor unrecognized component: %v", component)
}

func (tc *E2ETestFramework) waitForElasticsearchPods(retryInterval, timeout time.Duration) error {
	logger.Debugf("Waiting for %v", "elasticsearch")
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		options := metav1.ListOptions{
			LabelSelector: "component=elasticsearch",
		}
		pods, err := tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Debugf("Did not find elasticsearch pods %v", err)
				return false, nil
			}
			logger.Debugf("Error listing elasticsearch pods %v", err)
			return false, err
		}
		if len(pods.Items) == 0 {
			logger.Debugf("No elasticsearch pods found %v", pods)
			return false, nil
		}

		for _, pod := range pods.Items {
			for _, status := range pod.Status.ContainerStatuses {
				logger.Debugf("Checking status of %s.%s: %v", pod.Name, status.ContainerID, status.Ready)
				if !status.Ready {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

func (tc *E2ETestFramework) waitForDeployment(namespace, name string, retryInterval, timeout time.Duration) error {
	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := tc.KubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		replicas := int(*deployment.Spec.Replicas)
		if int(deployment.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		return false, nil
	})
}

func (tc *E2ETestFramework) WaitForCleanupCompletion(podlabels []string) {
	if err := tc.waitForClusterLoggingPodsCompletion(podlabels); err != nil {
		logger.Errorf("Cleanup completion error %v", err)
	}
}

func (tc *E2ETestFramework) waitForClusterLoggingPodsCompletion(podlabels []string) error {
	labels := strings.Join(podlabels, ",")
	logger.Infof("waiting for pods to complete with labels: %s", labels)
	labelSelector := fmt.Sprintf("component in (%s)", labels)
	options := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	return wait.Poll(defaultRetryInterval, defaultTimeout, func() (bool, error) {
		pods, err := tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Infof("Did not find pods %v", err)
				return false, nil
			}
			logger.Infof("Error listing pods %v", err)
			return false, err
		}
		if len(pods.Items) == 0 {
			logger.Infof("No pods found for label selection: %s", labels)
			return true, nil
		}
		logger.Debugf("%v pods still running", len(pods.Items))
		return false, nil
	})
}

func (tc *E2ETestFramework) SetupClusterLogging(componentTypes ...LogComponentType) error {
	tc.ClusterLogging = NewClusterLogging(componentTypes...)
	tc.LogStore = &ElasticLogStore{
		Framework: tc,
	}
	return tc.CreateClusterLogging(tc.ClusterLogging)
}

func (tc *E2ETestFramework) CreateClusterLogging(clusterlogging *cl.ClusterLogging) error {
	body, err := json.Marshal(clusterlogging)
	if err != nil {
		return err
	}
	logger.Debugf("Creating ClusterLogging: %s", string(body))
	result := tc.KubeClient.RESTClient().Post().
		RequestURI(clusterLoggingURI).
		SetHeader("Content-Type", "application/json").
		Body(body).
		Do()
	tc.AddCleanup(func() error {
		return tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/instance", clusterLoggingURI)).
			SetHeader("Content-Type", "application/json").
			Do().Error()
	})
	return result.Error()
}

func (tc *E2ETestFramework) CreateClusterLogForwarder(forwarder *logging.ClusterLogForwarder) error {
	body, err := json.Marshal(forwarder)
	if err != nil {
		return err
	}
	logger.Debugf("Creating ClusterLogForwarder: %s", string(body))
	result := tc.KubeClient.RESTClient().Post().
		RequestURI(clusterlogforwarderURI).
		SetHeader("Content-Type", "application/json").
		Body(body).
		Do()
	tc.AddCleanup(func() error {
		return tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/instance", clusterlogforwarderURI)).
			SetHeader("Content-Type", "application/json").
			Do().Error()
	})
	return result.Error()
}

func (tc *E2ETestFramework) Cleanup() {
	//allow caller to cleanup if unset (e.g script cleanup())
	logger.Infof("Running Cleanup....")
	doCleanup := strings.TrimSpace(os.Getenv("DO_CLEANUP"))
	if doCleanup == "" || strings.ToLower(doCleanup) == "true" {
		RunCleanupScript()
		logger.Debugf("Running %v e2e cleanup functions", len(tc.CleanupFns))
		for _, cleanup := range tc.CleanupFns {
			logger.Debug("Running an e2e cleanup function")
			if err := cleanup(); err != nil {
				logger.Debugf("Error during cleanup %v", err)
			}
		}
	}
}

func RunCleanupScript() {
	if value, found := os.LookupEnv("CLEANUP_CMD"); found {
		if strings.TrimSpace(value) == "" {
			logger.Info("No cleanup script provided")
			return
		}
		args := strings.Split(value, " ")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = nil
		result, err := cmd.CombinedOutput()
		logger.Infof("RunCleanupScript output: %s", string(result))
		logger.Infof("RunCleanupScript err: %v", err)
	}
}

//newKubeClient returns a client using the KUBECONFIG env var or incluster settings
func newKubeClient() (*kubernetes.Clientset, *rest.Config) {

	var config *rest.Config
	var err error
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, config
}

func (tc *E2ETestFramework) PodExec(namespace, name, container string, command []string) (string, error) {
	return oc.Exec().WithNamespace(namespace).Pod(name).Container(container).WithCmd(command[0], command[1:]...).Run()
}

func (tc *E2ETestFramework) CreatePipelineSecret(pwd, logStoreName, secretName string, otherData map[string][]byte) (secret *corev1.Secret, err error) {
	workingDir := fmt.Sprintf("/tmp/clo-test-%d", rand.Intn(10000))
	logger.Debugf("Generating Pipeline certificates for %q to %s", logStoreName, workingDir)
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		if err = os.MkdirAll(workingDir, 0766); err != nil {
			return nil, err
		}
	}
	if err = os.Setenv("WORKING_DIR", workingDir); err != nil {
		return nil, err
	}
	if err = k8shandler.GenerateCertificates(OpenshiftLoggingNS, pwd, logStoreName, workingDir); err != nil {
		return nil, err
	}
	data := map[string][]byte{
		"tls.key":       utils.GetWorkingDirFileContents("system.logging.fluentd.key"),
		"tls.crt":       utils.GetWorkingDirFileContents("system.logging.fluentd.crt"),
		"ca-bundle.crt": utils.GetWorkingDirFileContents("ca.crt"),
		"ca.key":        utils.GetWorkingDirFileContents("ca.key"),
	}
	for key, value := range otherData {
		data[key] = value
	}
	secret = k8shandler.NewSecret(
		secretName,
		OpenshiftLoggingNS,
		data,
	)
	logger.Debugf("Creating secret %s for logStore %s", secret.Name, logStoreName)
	if secret, err = tc.KubeClient.Core().Secrets(OpenshiftLoggingNS).Create(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

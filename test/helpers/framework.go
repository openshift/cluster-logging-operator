package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clolog "github.com/ViaQ/logerr/log"
	"github.com/onsi/ginkgo"
	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/certificates"
	k8shandler "github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

const (
	clusterLoggingURI      = "apis/logging.openshift.io/v1/namespaces/openshift-logging/clusterloggings"
	clusterlogforwarderURI = "apis/logging.openshift.io/v1/namespaces/openshift-logging/clusterlogforwarders"
	DefaultCleanUpTimeout  = 60.0 * 2
)

const (
	defaultRetryInterval      = 1 * time.Second
	defaultTimeout            = 10 * time.Minute
	DefaultWaitForLogsTimeout = 10 * time.Minute
	// Shorter timeout for when we are expecting that there will be no logs
	// since we will always go to the timeout in that case.a
	DefaultWaitForNoLogsTimeout = 1 * time.Minute
)

type LogStore interface {
	//ApplicationLogs returns app logs for a given log store
	ApplicationLogs(timeToWait time.Duration) (types.Logs, error)

	HasApplicationLogs(timeToWait time.Duration) (bool, error)

	HasInfraStructureLogs(timeToWait time.Duration) (bool, error)

	HasAuditLogs(timeToWait time.Duration) (bool, error)

	GrepLogs(expr string, timeToWait time.Duration) (string, error)

	RetrieveLogs() (map[string]string, error)

	ClusterLocalEndpoint() string
}

type E2ETestFramework struct {
	RestConfig     *rest.Config
	KubeClient     *kubernetes.Clientset
	ClusterLogging *cl.ClusterLogging
	CleanupFns     []func() error
	LogStores      map[string]LogStore
}

func NewE2ETestFramework() *E2ETestFramework {
	client, config := NewKubeClient()
	framework := &E2ETestFramework{
		RestConfig: config,
		KubeClient: client,
		LogStores:  make(map[string]LogStore, 4),
	}
	return framework
}

func (tc *E2ETestFramework) AddCleanup(fn func() error) {
	tc.CleanupFns = append(tc.CleanupFns, fn)
}

func (tc *E2ETestFramework) DeployLogGenerator() error {
	namespace := tc.CreateTestNamespace()
	return tc.DeployLogGeneratorWithNamespace(namespace)
}

func (tc *E2ETestFramework) DeployLogGeneratorWithNamespace(namespace string) error {
	opts := metav1.CreateOptions{}
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
	clolog.Info("Deploying LogGenerator to namespace", "deployment name", deployment.Name, "namespace", deployment.Namespace)
	deployment, err := tc.KubeClient.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, opts)
	if err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), deployment.Name, opts)
	})
	return tc.waitForDeployment(namespace, "log-generator", defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) DeployLogGeneratorWithNamespaceAndLabels(namespace string, labels map[string]string) error {
	err := tc.DeployLogGeneratorWithNamespace(namespace)
	if err != nil {
		return err
	}
	for k, v := range labels {
		if _, err2 := oc.Literal().From("oc label pod -n %s --all %s=%s", namespace, k, v).Run(); err != nil {
			return err2
		}
	}
	return err
}

func (tc *E2ETestFramework) DeployJsonLogGenerator(vals map[string]string) (string, string, error) {
	namespace := tc.CreateTestNamespace()
	pycode := `
import time,json,sys,datetime
%s
i=0
while True:
  i=i+1
  ts=time.time()
  data={
	"timestamp"   :datetime.datetime.fromtimestamp(ts).strftime('%%Y-%%m-%%d %%H:%%M:%%S'),
	"index"       :i,
  }
  set_vals()
  print json.dumps(data)
  sys.stdout.flush()
  time.sleep(1)
`
	setVals := `
def set_vals():
  pass

`
	if len(vals) != 0 {
		setVals = "def set_vals():\n"
		for k, v := range vals {
			//...  data["key"]="value"
			setVals += fmt.Sprintf("  data[\"%s\"]=\"%s\"\n", k, v)
		}
		setVals += "\n"
	}
	container := corev1.Container{
		Name:            "log-generator",
		Image:           "centos:centos7",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args:            []string{"python2", "-c", fmt.Sprintf(pycode, setVals)},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
	}
	deployment := k8shandler.NewDeployment("log-generator", namespace, "log-generator", "test", podSpec)
	clolog.Info("Deploying deployment to namespace", "deployment", deployment.Name, "namespace", deployment.Namespace)
	deployment, err := tc.KubeClient.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return "", "", err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{})
	})
	err = tc.waitForDeployment(namespace, "log-generator", defaultRetryInterval, defaultTimeout)
	if err == nil {
		podName, _ := oc.Get().WithNamespace(namespace).Pod().OutputJsonpath("{.items[0].metadata.name}").Run()
		return namespace, podName, nil

	}
	return "", "", err
}

func (tc *E2ETestFramework) CreateTestNamespace() string {
	opts := metav1.CreateOptions{}
	name := fmt.Sprintf("clo-test-%d", rand.Intn(10000))
	if value, found := os.LookupEnv("GENERATOR_NS"); found {
		name = value
	}
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := tc.KubeClient.CoreV1().Namespaces().Create(context.TODO(), namespace, opts)
	if err != nil && !errors.IsAlreadyExists(err) {
		clolog.Error(err, "Error:")
	}
	return name
}

func (tc *E2ETestFramework) WaitFor(component LogComponentType) error {
	switch component {
	case ComponentTypeVisualization:
		return tc.waitForDeployment(OpenshiftLoggingNS, "kibana", defaultRetryInterval, defaultTimeout)
	case ComponentTypeCollector:
		clolog.V(3).Info("Waiting for ", "component", component)
		return tc.waitForFluentDaemonSet(defaultRetryInterval, defaultTimeout)
	case ComponentTypeStore:
		return tc.waitForElasticsearchPods(defaultRetryInterval, defaultTimeout)
	}
	return fmt.Errorf("Unable to waitfor unrecognized component: %v", component)
}

func (tc *E2ETestFramework) waitForFluentDaemonSet(retryInterval, timeout time.Duration) error {
	// daemonset should have pods running and available on all the nodes for maxtimes * retryInterval
	maxtimes := 5
	times := 0
	return wait.PollImmediate(retryInterval, timeout, func() (bool, error) {
		numUnavail, err := oc.Literal().From("oc -n openshift-logging get daemonset/fluentd -o jsonpath={.status.NumberUnavailable}").Run()
		if err == nil {
			if numUnavail == "" {
				numUnavail = "0"
			}
			value, err := strconv.Atoi(strings.TrimSpace(numUnavail))
			if err != nil {
				times = 0
				return false, err
			}
			if value == 0 {
				times++
			} else {
				times = 0
			}
			if times == maxtimes {
				return true, nil
			}
		}
		return false, nil
	})
}

func (tc *E2ETestFramework) waitForElasticsearchPods(retryInterval, timeout time.Duration) error {
	clolog.V(3).Info("Waiting for elasticsearch")
	return wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		options := metav1.ListOptions{
			LabelSelector: "component=elasticsearch",
		}
		pods, err := tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(context.TODO(), options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				clolog.V(2).Error(err, "Did not find elasticsearch pods")
				return false, nil
			}
			clolog.Error(err, "Error listing elasticsearch pods")
			return false, nil
		}
		numberOfPods := len(pods.Items)
		if numberOfPods == 0 {
			clolog.V(2).Info("No elasticsearch pods found ", "pods", pods)
			return false, nil
		}
		containersReadyCount := 0
		containersNotReadyCount := 0
		for _, pod := range pods.Items {
			for _, status := range pod.Status.ContainerStatuses {
				clolog.V(3).Info("Checking status of", "PodName", pod.Name, "ContainerID", status.ContainerID, "status", status.Ready)
				if status.Ready {
					containersReadyCount++
				} else {
					containersNotReadyCount++
				}
			}
		}
		if containersReadyCount == 0 || containersNotReadyCount > 0 {
			clolog.V(3).Info("elasticsearch containers are not ready", "pods", numberOfPods, "ready containers", containersReadyCount, "not ready containers", containersNotReadyCount)
			return false, nil
		}

		return true, nil
	})
}

func (tc *E2ETestFramework) waitForDeployment(namespace, name string, retryInterval, timeout time.Duration) error {
	return wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := tc.KubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			clolog.Error(err, "Error trying to retrieve deployment")
			return false, nil
		}
		replicas := int(*deployment.Spec.Replicas)
		if int(deployment.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		return false, nil
	})
}

func (tc *E2ETestFramework) WaitForCleanupCompletion(namespace string, podlabels []string) {
	if err := tc.waitForClusterLoggingPodsCompletion(namespace, podlabels); err != nil {
		clolog.Error(err, "Cleanup completion error")
	}
}

func (tc *E2ETestFramework) waitForClusterLoggingPodsCompletion(namespace string, podlabels []string) error {
	labels := strings.Join(podlabels, ",")
	clolog.Info("waiting for pods to complete with labels in namespace:", "labels", labels, "namespace", namespace)
	labelSelector := fmt.Sprintf("component in (%s)", labels)
	options := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	return wait.PollImmediate(defaultRetryInterval, defaultTimeout, func() (bool, error) {
		pods, err := tc.KubeClient.CoreV1().Pods(namespace).List(context.TODO(), options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				clolog.Error(err, "Did not find pods")
				return false, nil
			}
			clolog.Error(err, "Error listing pods ")
			return false, nil
		}
		if len(pods.Items) == 0 {
			clolog.Info("No pods found for label selection", "labels", labels)
			return true, nil
		}
		clolog.V(3).Info("pods still running", "num", len(pods.Items))
		return false, nil
	})
}

func (tc *E2ETestFramework) waitForStatefulSet(namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := tc.KubeClient.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			clolog.Error(err, "Error Getting StatfuleSet")
			return false, nil
		}
		replicas := int(*deployment.Spec.Replicas)
		if int(deployment.Status.ReadyReplicas) == replicas {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (tc *E2ETestFramework) SetupClusterLogging(componentTypes ...LogComponentType) error {
	tc.ClusterLogging = NewClusterLogging(componentTypes...)
	tc.LogStores["elasticsearch"] = &ElasticLogStore{
		Framework: tc,
	}
	return tc.CreateClusterLogging(tc.ClusterLogging)
}

func (tc *E2ETestFramework) CreateClusterLogging(clusterlogging *cl.ClusterLogging) error {
	body, err := json.Marshal(clusterlogging)
	if err != nil {
		return err
	}
	clolog.V(3).Info("Creating ClusterLogging:", "ClusterLogging", string(body))
	result := tc.KubeClient.RESTClient().Post().
		RequestURI(clusterLoggingURI).
		SetHeader("Content-Type", "application/json").
		Body(body).
		Do(context.TODO())
	tc.AddCleanup(func() error {
		return tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/instance", clusterLoggingURI)).
			SetHeader("Content-Type", "application/json").
			Do(context.TODO()).Error()
	})
	return result.Error()
}

func (tc *E2ETestFramework) CreateClusterLogForwarder(forwarder *logging.ClusterLogForwarder) error {
	body, err := json.Marshal(forwarder)
	if err != nil {
		return err
	}
	clolog.V(3).Info("Creating ClusterLogForwarder", "ClusterLogForwarder", string(body))
	result := tc.KubeClient.RESTClient().Post().
		RequestURI(clusterlogforwarderURI).
		SetHeader("Content-Type", "application/json").
		Body(body).
		Do(context.TODO())
	tc.AddCleanup(func() error {
		return tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/instance", clusterlogforwarderURI)).
			SetHeader("Content-Type", "application/json").
			Do(context.TODO()).Error()
	})
	return result.Error()
}

func (tc *E2ETestFramework) Cleanup() {
	if ginkgo.CurrentGinkgoTestDescription().Failed {
		//allow caller to cleanup if unset (e.g script cleanup())
		clolog.Info("Running Cleanup script ....")
		doCleanup := strings.TrimSpace(os.Getenv("DO_CLEANUP"))
		if doCleanup == "" || strings.ToLower(doCleanup) == "true" {
			RunCleanupScript()
		}
	} else {
		clolog.Info("Not Running Cleanup script")
	}
	clolog.V(3).Info("Running e2e cleanup functions, ", "number", len(tc.CleanupFns))
	for _, cleanup := range tc.CleanupFns {
		clolog.V(3).Info("Running an e2e cleanup function")
		if err := cleanup(); err != nil {
			clolog.V(2).Info("Error during cleanup ", "error", err)
		}
	}
	tc.CleanFluentDBuffers()
}

func RunCleanupScript() {
	if value, found := os.LookupEnv("CLEANUP_CMD"); found {
		if strings.TrimSpace(value) == "" {
			clolog.Info("No cleanup script provided")
			return
		}
		args := strings.Split(value, " ")
		// #nosec G204
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = nil
		result, err := cmd.CombinedOutput()
		clolog.Info("RunCleanupScript output: ", "output", string(result))
		clolog.Info("RunCleanupScript err: ", "error", err)
	}
}

func (tc *E2ETestFramework) CleanFluentDBuffers() {
	h := corev1.HostPathDirectory
	p := true
	spec := &v1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clean-fluentd-buffers",
			Namespace: "default",
		},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "clean-fluentd-buffers",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "clean-fluentd-buffers",
					},
				},
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "clean-fluentd-buffers",
							Image: "centos:centos7",
							Args:  []string{"sh", "-c", "rm -rf /fluentd-buffers/** || rm /logs/audit/audit.log.pos || rm /logs/kube-apiserver/audit.log.pos || rm /logs/es-containers.log.pos"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &p,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fluentd-buffers",
									MountPath: "/fluentd-buffers",
								},
								{
									Name:      "logs",
									MountPath: "/logs",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "pause",
							Image: "centos:centos7",
							Args:  []string{"sh", "-c", "echo done!!!!"},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "fluentd-buffers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/fluentd",
									Type: &h,
								},
							},
						},
						{
							Name: "logs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log",
									Type: &h,
								},
							},
						},
					},
				},
			},
		},
	}
	ds, err := tc.KubeClient.AppsV1().DaemonSets("default").Create(context.TODO(), spec, metav1.CreateOptions{})
	if err != nil {
		clolog.Error(err, "Could not create DaemonSet for cleaning fluentd buffers.")
		return
	} else {
		clolog.Info("DaemonSet to clean fluent buffers created")
	}
	_ = wait.PollImmediate(time.Second*10, time.Minute*5, func() (bool, error) {
		desired, err2 := oc.Get().Resource("daemonset", "clean-fluentd-buffers").WithNamespace("default").OutputJsonpath("{.status.desiredNumberScheduled}").Run()
		if err2 != nil {
			return false, nil
		}
		current, err2 := oc.Get().Resource("daemonset", "clean-fluentd-buffers").WithNamespace("default").OutputJsonpath("{.status.currentNumberScheduled}").Run()
		if err2 != nil {
			return false, nil
		}
		if current == desired {
			return true, nil
		}
		return false, nil
	})
	err = tc.KubeClient.AppsV1().DaemonSets(ds.GetNamespace()).Delete(context.TODO(), ds.GetName(), metav1.DeleteOptions{})
	if err != nil {
		clolog.Error(err, "Could not delete DaemonSet for cleaning fluentd buffers.")
	} else {
		clolog.Info("DaemonSet to clean fluent buffers deleted")
	}
}

//NewKubeClient returns a client using the KUBECONFIG env var or incluster settings
func NewKubeClient() (*kubernetes.Clientset, *rest.Config) {

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

func (tc *E2ETestFramework) CreatePipelineSecret(pwd, logStoreName, secretName string, otherData map[string][]byte) (*corev1.Secret, error) {
	workingDir := fmt.Sprintf("/tmp/clo-test-%d", rand.Intn(10000))
	clolog.V(3).Info("Generating Pipeline certificates for Log Store to working dir", "logStoreName", logStoreName, "workingDir", workingDir)
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		if err = os.MkdirAll(workingDir, 0766); err != nil {
			return nil, err
		}
	}
	if err := os.Setenv("WORKING_DIR", workingDir); err != nil {
		return nil, err
	}
	scriptsDir := fmt.Sprintf("%s/scripts", pwd)
	if err, _, _ := certificates.GenerateCertificates(OpenshiftLoggingNS, scriptsDir, logStoreName, workingDir); err != nil {
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

	sOpts := metav1.CreateOptions{}
	secret := k8shandler.NewSecret(
		secretName,
		OpenshiftLoggingNS,
		data,
	)
	clolog.V(3).Info("Creating secret for logStore ", "secret", secret.Name, "logStoreName", logStoreName)
	newSecret, err := tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Create(context.TODO(), secret, sOpts)
	if err == nil {
		return newSecret, nil
	}
	if errors.IsAlreadyExists(err) {
		sOpts := metav1.UpdateOptions{}
		updatedSecret, err := tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Update(context.TODO(), secret, sOpts)
		if err == nil {
			return updatedSecret, nil
		}
	}

	return nil, err
}

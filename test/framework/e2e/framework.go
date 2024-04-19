package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	commonlog "github.com/openshift/cluster-logging-operator/test/framework/common/log"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e/receivers/elasticsearch"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	"golang.org/x/sync/errgroup"

	"github.com/openshift/cluster-logging-operator/test/runtime"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers"

	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	cl "github.com/openshift/cluster-logging-operator/api/logging/v1"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	k8shandler "github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

func init() {
	var verbosity = 0
	if level, found := os.LookupEnv("LOG_LEVEL"); found {
		if i, err := strconv.Atoi(level); err == nil {
			verbosity = i
		}
	}
	failureLogger, delayedWriter := commonlog.NewLogger("e2e-framework", verbosity)
	clolog.SetLogger(failureLogger)
	delayedLogWriter = delayedWriter
}

const (
	clusterLoggingURI      = "apis/logging.openshift.io/v1/namespaces/%s/clusterloggings"
	clusterlogforwarderURI = "apis/logging.openshift.io/v1/namespaces/%s/clusterlogforwarders"
	DefaultCleanUpTimeout  = 60.0 * 5

	defaultRetryInterval      = 1 * time.Second
	defaultTimeout            = 5 * time.Minute
	DefaultWaitForLogsTimeout = 5 * time.Minute
)

var (
	delayedLogWriter *commonlog.BufferedLogWriter
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

func (tc *E2ETestFramework) DeployLogGenerator() (string, error) {
	namespace := tc.CreateTestNamespace()
	return namespace, tc.DeployLogGeneratorWithNamespaceName(namespace, "log-generator", NewDefaultLogGeneratorOptions())
}

func (tc *E2ETestFramework) DeployCURLLogGenerator(endpoint string) (string, error) {
	namespace := tc.CreateTestNamespace()
	return namespace, tc.DeployCURLLogGeneratorWithNamespaceAndEndpoint(namespace, endpoint)
}

type LogGeneratorOptions struct {
	Count          int
	Delay          time.Duration
	Message        string
	ContainerCount int
	Labels         map[string]string
}

func NewDefaultLogGeneratorOptions() LogGeneratorOptions {
	return LogGeneratorOptions{Count: 1000, Delay: 0, Message: "My life is my message", ContainerCount: 1, Labels: map[string]string{}}
}

func (tc *E2ETestFramework) DeployLogGeneratorWithNamespaceName(namespace, name string, options LogGeneratorOptions) error {
	pod := runtime.NewMultiContainerLogGenerator(namespace, name, options.Count, options.Delay, options.Message, options.ContainerCount, options.Labels)
	clolog.Info("Checking SA for LogGenerator", "Deployment name", pod.Name, "namespace", namespace)
	if err := tc.WaitForResourceCondition(namespace, "serviceaccount", "default", "", "{}", 10, func(string) (bool, error) { return true, nil }); err != nil {
		return err
	}
	clolog.Info("Deploying LogGenerator to namespace", "Deployment name", pod.Name, "namespace", namespace)
	opts := metav1.CreateOptions{}
	pod, err := tc.KubeClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, opts)
	if err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, opts)
	})
	return client.Get().WaitFor(pod, client.PodRunning)
}

// DeploySocat will deploy pod with socat software
func (tc *E2ETestFramework) DeploySocat(namespace, name, forwarderName string, options LogGeneratorOptions) error {
	pod := runtime.NewSocatPod(namespace, name, forwarderName, options.Labels)
	if err := tc.WaitForResourceCondition(namespace, "serviceaccount", "default", "", "{}", 10, func(string) (bool, error) { return true, nil }); err != nil {
		return err
	}
	opts := metav1.CreateOptions{}
	pod, err := tc.KubeClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, opts)
	if err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, opts)
	})
	return client.Get().WaitFor(pod, client.PodRunning)
}

func (tc *E2ETestFramework) DeployLogGeneratorWithNamespace(namespace, name string, options LogGeneratorOptions) error {
	return tc.DeployLogGeneratorWithNamespaceName(namespace, name, options)
}

func (tc *E2ETestFramework) DeployJsonLogGenerator(vals, labels map[string]string) (string, string, error) {
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
	for k, v := range labels {
		deployment.Spec.Template.Labels[k] = v
	}
	clolog.Info("Deploying Deployment to namespace", "Deployment", deployment.Name, "namespace", deployment.Namespace)
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

func (tc *E2ETestFramework) DeployCURLLogGeneratorWithNamespaceAndEndpoint(namespace, endpoint string) error {
	if err := tc.WaitForResourceCondition(namespace, "serviceaccount", "default", "", "{}", 10, func(string) (bool, error) { return true, nil }); err != nil {
		return err
	}
	pod := runtime.NewCURLLogGenerator(namespace, "log-generator", endpoint, 0, 0, "My life is my message")
	clolog.Info("Deploying LogGenerator to namespace", "deployment name", pod.Name, "namespace", namespace)
	opts := metav1.CreateOptions{}
	pod, err := tc.KubeClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, opts)
	if err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, opts)
	})
	return client.Get().WaitFor(pod, client.PodRunning)
}

func (tc *E2ETestFramework) CreateTestNamespace() string {
	return tc.CreateTestNamespaceWithPrefix("clo-test")
}

func (tc *E2ETestFramework) CreateTestNamespaceWithPrefix(prefix string) string {
	opts := metav1.CreateOptions{}
	name := fmt.Sprintf("%s-%d", prefix, rand.Intn(10000)) //nolint:gosec
	if value, found := os.LookupEnv("GENERATOR_NS"); found {
		name = value
	} else {
		tc.AddCleanup(func() error {
			opts := metav1.DeleteOptions{}
			return tc.KubeClient.CoreV1().Namespaces().Delete(context.TODO(), name, opts)
		})
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

func (tc *E2ETestFramework) DeployComponents(componentTypes ...helpers.LogComponentType) error {
	for _, comp := range componentTypes {
		switch comp {
		case helpers.ComponentTypeReceiverElasticsearchRHManaged:
			receiver := elasticsearch.NewManagedElasticsearch(tc)
			if err := receiver.Deploy(); err != nil {
				return err
			}
			tc.LogStores[elasticsearch.ManagedLogStore] = receiver
		case helpers.ComponentTypeCollectorVector:
			clf := runtime.NewClusterLogForwarder()
			clf.Name = "mycollector"
			runtime.NewClusterLogForwarderBuilder(clf).
				FromInput(logging.InputNameApplication).
				AndInput(logging.InputNameInfrastructure).
				AndInput(logging.InputNameAudit).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Type = logging.OutputTypeElasticsearch
					spec.URL = "https://elasticsearch:9200"
					spec.Secret = &logging.OutputSecretSpec{
						Name: clf.Name,
					}
				}, elasticsearch.ManagedLogStore)
			if err := tc.CreateServiceAccountAndAuthorizeFor(clf); err != nil {
				return err
			}
			if err := tc.CreateClusterLogForwarder(clf); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tc *E2ETestFramework) Client() *kubernetes.Clientset {
	return tc.KubeClient
}

func (tc *E2ETestFramework) SetupClusterLogging(componentTypes ...helpers.LogComponentType) (err error) {
	return tc.DeployComponents()
}

func (tc *E2ETestFramework) CreateClusterLogging(clusterlogging *cl.ClusterLogging) error {
	body, err := json.Marshal(clusterlogging)
	if err != nil {
		return err
	}
	deleteCL := func() error {
		return tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/instance", fmt.Sprintf(clusterLoggingURI, clusterlogging.Namespace))).
			SetHeader("Content-Type", "application/json").
			Do(context.TODO()).Error()
	}
	createCL := func() error {
		clolog.Info("Creating ClusterLogging:", "ClusterLogging", string(body))
		return tc.KubeClient.RESTClient().Post().
			RequestURI(fmt.Sprintf(clusterLoggingURI, clusterlogging.Namespace)).
			SetHeader("Content-Type", "application/json").
			Body(body).
			Do(context.TODO()).Error()
	}
	tc.AddCleanup(deleteCL)
	err = createCL()
	if apierrors.IsAlreadyExists(err) {
		clolog.Info("clusterlogging instance already exists. Attempting to re-deploy...")
		if err = deleteCL(); err != nil {
			clolog.Error(err, "failed deleting clusterlogging instance")
			return err
		}
		err = createCL()
	}
	return err
}

func (tc *E2ETestFramework) CreateClusterLogForwarder(forwarder *logging.ClusterLogForwarder) error {
	body, err := json.Marshal(forwarder)
	if err != nil {
		return err
	}
	deleteCLF := func() error {
		return tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/instance", fmt.Sprintf(clusterlogforwarderURI, forwarder.Namespace))).
			SetHeader("Content-Type", "application/json").
			Do(context.TODO()).Error()
	}
	tc.AddCleanup(deleteCLF)
	clolog.Info("Creating ClusterLogForwarder", "ClusterLogForwarder", string(body))
	createCLF := func() rest.Result {
		return tc.KubeClient.RESTClient().Post().
			RequestURI(fmt.Sprintf(clusterlogforwarderURI, forwarder.Namespace)).
			SetHeader("Content-Type", "application/json").
			Body(body).
			Do(context.TODO())
	}
	result := createCLF()
	if err := result.Error(); err != nil && apierrors.IsAlreadyExists(err) {
		clolog.Info("clusterlogforwarder instance already exists. Removing and trying to recreate...")
		if err := deleteCLF(); err != nil {
			return err
		}
		result = createCLF()
	}

	return result.Error()
}

func DoCleanup() bool {
	doCleanup := strings.TrimSpace(os.Getenv("DO_CLEANUP"))
	clolog.Info("Running Cleanup script ....", "DO_CLEANUP", doCleanup)
	return doCleanup == "" || strings.ToLower(doCleanup) == "true"
}

func (tc *E2ETestFramework) Cleanup() {
	if g, ok := test.GinkgoCurrentTest(); ok && g.Failed {
		clolog.Info("Test failed", "TestText", g.FullTestText)
		//allow caller to cleanup if unset (e.g script cleanup())
		if DoCleanup() {
			RunCleanupScript()
		} else {
			return
		}
	} else {
		clolog.V(1).Info("Test passed. Skipping artifacts gathering")
	}
	clolog.Info("Running e2e cleanup functions, ", "number", len(tc.CleanupFns))
	for _, cleanup := range tc.CleanupFns {
		clolog.V(5).Info("Running an e2e cleanup function")
		if err := cleanup(); err != nil {
			if !apierrors.IsNotFound(err) {
				clolog.V(2).Info("Error during cleanup ", "error", err)
			}
		}
	}
	tc.CleanupFns = [](func() error){}
	if tc.ClusterLogging != nil && tc.ClusterLogging.Spec.Collection != nil &&
		(tc.ClusterLogging.Spec.Collection.Type == logging.LogCollectionTypeFluentd ||
			tc.ClusterLogging.Spec.Collection.Logs != nil && tc.ClusterLogging.Spec.Collection.Logs.Type == logging.LogCollectionTypeFluentd) {
		tc.CleanFluentDBuffers()
	}
	if g, ok := test.GinkgoCurrentTest(); ok && g.Failed {
		delayedLogWriter.Flush()
	}
}

func RunCleanupScript() {
	if value, found := os.LookupEnv("CLEANUP_CMD"); found {
		if strings.TrimSpace(value) == "" {
			clolog.Info("No cleanup script provided")
			return
		}
		clolog.Info("Script", "CLEANUP_CMD", value)
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
			Name:      "clean-buffers",
			Namespace: "default",
		},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "clean-buffers",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "clean-buffers",
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
							Name:  "clean-buffers",
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
	_ = wait.PollUntilContextTimeout(context.TODO(), time.Second*10, time.Minute*5, true, func(cxt context.Context) (done bool, err error) {
		desired, err2 := oc.Get().Resource("daemonset", "clean-buffers").WithNamespace("default").OutputJsonpath("{.status.desiredNumberScheduled}").Run()
		if err2 != nil {
			return false, nil
		}
		current, err2 := oc.Get().Resource("daemonset", "clean-buffers").WithNamespace("default").OutputJsonpath("{.status.currentNumberScheduled}").Run()
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

// NewKubeClient returns a client using the KUBECONFIG env var or incluster settings
func NewKubeClient() (*kubernetes.Clientset, *rest.Config) {
	config, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, config
}

func (tc *E2ETestFramework) PodExec(namespace, pod, container string, command []string) (string, error) {
	return oc.Exec().WithNamespace(namespace).Pod(pod).Container(container).WithCmd(command[0], command[1:]...).Run()
}

func (tc *E2ETestFramework) CreatePipelineSecret(logStoreName, secretName string, otherData map[string][]byte) (*corev1.Secret, error) {
	ca := certificate.NewCA(nil, "Root CA") // Self-signed CA
	serverCert := certificate.NewCert(ca, "", logStoreName, fmt.Sprintf("%s.%s.svc", logStoreName, constants.OpenshiftNS))

	data := map[string][]byte{
		"tls.key":       serverCert.PrivateKeyPEM(),
		"tls.crt":       serverCert.CertificatePEM(),
		"ca-bundle.crt": ca.CertificatePEM(),
		"ca.key":        ca.PrivateKeyPEM(),
	}
	for key, value := range otherData {
		data[key] = value
	}

	sOpts := metav1.CreateOptions{}
	secret := k8shandler.NewSecret(
		secretName,
		constants.OpenshiftNS,
		data,
	)
	clolog.V(3).Info("Creating secret for logStore ", "secret", secret.Name, "logStoreName", logStoreName)
	newSecret, err := tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Create(context.TODO(), secret, sOpts)
	if err == nil {
		return newSecret, nil
	}
	if errors.IsAlreadyExists(err) {
		sOpts := metav1.UpdateOptions{}
		updatedSecret, err := tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Update(context.TODO(), secret, sOpts)
		if err == nil {
			return updatedSecret, nil
		}
	}

	return nil, err
}

// CLF depends on CL, so sync the creator goroutines
func RecreateClClfAsync(g *errgroup.Group, c *client.Test, cl *logging.ClusterLogging, clf *logging.ClusterLogForwarder) {
	ch := make(chan struct{}, 1)
	g.Go(func() error {
		defer func() { ch <- struct{}{} }()
		return c.Recreate(cl)
	})
	g.Go(func() error {
		defer func() { close(ch) }()
		<-ch
		return c.Recreate(clf)
	})
}

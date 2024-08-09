package e2e

import (
	"context"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	commonlog "github.com/openshift/cluster-logging-operator/test/framework/common/log"
	"math/rand"
	"os"
	"os/exec"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"

	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	cl "github.com/openshift/cluster-logging-operator/api/logging/v1"
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
	DefaultCleanUpTimeout = 60.0 * 5

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
	Test           *client.Test
}

func NewE2ETestFramework() *E2ETestFramework {
	kubeClient, config := NewKubeClient()
	framework := &E2ETestFramework{
		RestConfig: config,
		KubeClient: kubeClient,
		LogStores:  make(map[string]LogStore, 4),
		Test:       client.NewTest(),
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
	pod := testruntime.NewMultiContainerLogGenerator(namespace, name, options.Count, options.Delay, options.Message, options.ContainerCount, options.Labels)
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
	pod := testruntime.NewSocatPod(namespace, name, forwarderName, options.Labels)
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

func (tc *E2ETestFramework) DeployCURLLogGeneratorWithNamespaceAndEndpoint(namespace, endpoint string) error {
	pod := testruntime.NewCURLLogGenerator(namespace, "log-generator", endpoint, 0, 0, "My life is my message")
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
	name := fmt.Sprintf("%s-%d", prefix, rand.Intn(10000)) //nolint:gosec
	return tc.CreateNamespace(name)
}

func (tc *E2ETestFramework) CreateNamespace(name string) string {
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

	if err := tc.Test.Recreate(namespace); err != nil {
		clolog.Error(err, "Error")
		os.Exit(1)
	}
	clolog.V(1).Info("Created namespace", "namespace", name)
	return name
}

func (tc *E2ETestFramework) Client() *kubernetes.Clientset {
	return tc.KubeClient
}

// Create (re)creates an object
func (tc *E2ETestFramework) Create(obj crclient.Object) error {

	// Test round trip serialization
	body, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(body, obj); err != nil {
		return err
	}

	tc.AddCleanup(func() error {
		return tc.Test.Delete(obj)
	})
	clolog.Info("Creating object", "obj", string(body))
	return tc.Test.Recreate(obj)
}

func (tc *E2ETestFramework) CreateObservabilityClusterLogForwarder(forwarder *obs.ClusterLogForwarder) error {
	return tc.Create(forwarder)
}

func DoCleanup() bool {
	doCleanup := strings.TrimSpace(os.Getenv("DO_CLEANUP"))
	clolog.Info("Running Cleanup script ....", "DO_CLEANUP", doCleanup)
	return doCleanup == "" || strings.ToLower(doCleanup) == "true"
}

func (tc *E2ETestFramework) Cleanup() {
	if g, ok := test.GinkgoCurrentTest(); ok && g.Failed {
		defer delayedLogWriter.Flush()
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

func (tc *E2ETestFramework) CreatePipelineSecret(namespace, logStoreName, secretName string, otherData map[string][]byte) (*corev1.Secret, error) {
	ca := certificate.NewCA(nil, "Self-signed Root CA") // Self-signed CA
	serverCert := certificate.NewCert(ca, "Server Test CA", logStoreName, fmt.Sprintf("%s.%s.svc", logStoreName, namespace))
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
	secret := runtime.NewSecret(
		namespace,
		secretName,
		data,
	)
	clolog.V(3).Info("Creating secret for logStore ", "secret", secret.Name, "logStoreName", logStoreName)
	newSecret, err := tc.KubeClient.CoreV1().Secrets(namespace).Create(context.TODO(), secret, sOpts)
	if err == nil {
		return newSecret, nil
	}
	if errors.IsAlreadyExists(err) {
		sOpts := metav1.UpdateOptions{}
		updatedSecret, err := tc.KubeClient.CoreV1().Secrets(namespace).Update(context.TODO(), secret, sOpts)
		if err == nil {
			return updatedSecret, nil
		}
	}

	return nil, err
}

package functional

import (
	"encoding/base64"
	"fmt"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/certificates"

	yaml "sigs.k8s.io/yaml"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/components/fluentd"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	applicationLog     = "application"
	auditLog           = "audit"
	ovnAuditLog        = "ovn"
	k8sAuditLog        = "k8s"
	infraLog           = "infra"
	OpenshiftAuditLog  = "openshift-audit-logs"
	ApplicationLogFile = "/tmp/app-logs"
	FunctionalNodeName = "functional-test-node"
)

var fluentdLogPath = map[string]string{
	applicationLog:    "/var/log/containers",
	auditLog:          "/var/log/audit",
	ovnAuditLog:       "/var/log/ovn",
	OpenshiftAuditLog: "/var/log/oauth-apiserver",
	k8sAuditLog:       "/var/log/kube-apiserver",
}

var outputLogFile = map[string]map[string]string{
	logging.OutputTypeFluentdForward: {
		applicationLog: ApplicationLogFile,
		auditLog:       "/tmp/audit-logs",
		ovnAuditLog:    "/tmp/audit-logs",
		k8sAuditLog:    "/tmp/audit-logs",
		infraLog:       "/tmp/infra-logs",
	},
	logging.OutputTypeSyslog: {
		applicationLog: "/var/log/infra.log",
		auditLog:       "/var/log/infra.log",
		k8sAuditLog:    "/var/log/infra.log",
		ovnAuditLog:    "/var/log/infra.log",
		infraLog:       "/var/log/infra.logs",
	},
}

var (
	maxDuration          time.Duration
	defaultRetryInterval time.Duration

	//Timestamp = 2021-07-06T08:26:58.687Z
	OVNLogTemplate            = "%s|00004|acl_log(ovn_pinctrl0)|INFO|name=verify-audit-logging_deny-all, verdict=drop"
	KubeAuditLogTemplate      = `{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"a6299d35-5759-4f67-9bed-2b962cf21cf3","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/openshift-kube-storage-version-migrator/serviceaccounts/kube-storage-version-migrator-sa","verb":"get","user":{"username":"system:serviceaccount:openshift-kube-storage-version-migrator-operator:kube-storage-version-migrator-operator","uid":"d40a1a15-8b96-4ffa-a56b-5a834583532e","groups":["system:serviceaccounts","system:serviceaccounts:openshift-kube-storage-version-migrator-operator","system:authenticated"]},"sourceIPs":["10.128.0.16"],"userAgent":"cluster-kube-storage-version-migrator-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"serviceaccounts","namespace":"openshift-kube-storage-version-migrator","name":"kube-storage-version-migrator-sa","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"%s","stageTimestamp":"%s","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-kube-storage-version-migrator-operator:kube-storage-version-migrator-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:kube-storage-version-migrator-operator\" of ClusterRole \"cluster-admin\" to ServiceAccount \"kube-storage-version-migrator-operator/openshift-kube-storage-version-migrator-operator\""}}`
	OpenShiftAuditLogTemplate = `{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"19f44b1a-e4fb-4c9a-bc2f-068dc94be8fb","stage":"ResponseComplete","requestURI":"/","verb":"get","user":{"username":"system:anonymous","groups":["system:unauthenticated"]},"sourceIPs":["10.128.0.1"],"userAgent":"Go-http-client/1.1","responseStatus":{"metadata":{},"status":"Failure","reason":"Forbidden","code":403},"requestReceivedTimestamp":"%s","stageTimestamp":"%s","annotations":{"authorization.k8s.io/decision":"forbid","authorization.k8s.io/reason":""}}`
)

func NewKubeAuditLog(eventTime time.Time) string {
	now := CRIOTime(eventTime)
	return fmt.Sprintf(KubeAuditLogTemplate, now, now)
}

func NewAuditHostLog(eventTime time.Time) string {
	now := fmt.Sprintf("%.3f", float64(eventTime.UnixNano())/float64(time.Second))
	return fmt.Sprintf(`type=DAEMON_START msg=audit(%s:2914): op=start ver=3.0 format=enriched kernel=4.18.0-240.15.1.el8_3.x86_64 auid=4294967295 pid=1396 uid=0 ses=4294967295 subj=system_u:system_r:auditd_t:s0 res=successAUID="unset" UID="root"`, now)
}
func NewOVNAuditLog(eventTime time.Time) string {
	now := CRIOTime(eventTime)
	return fmt.Sprintf(OVNLogTemplate, now)
}

type receiverBuilder func(f *FluentdFunctionalFramework, b *runtime.PodBuilder, output logging.OutputSpec) error

//FluentdFunctionalFramework deploys stand alone fluentd with the fluent.conf as generated by input ClusterLogForwarder CR
type FluentdFunctionalFramework struct {
	Name              string
	Namespace         string
	Conf              string
	image             string
	labels            map[string]string
	Forwarder         *logging.ClusterLogForwarder
	Test              *client.Test
	Pod               *corev1.Pod
	fluentContainerId string
	receiverBuilders  []receiverBuilder
	closeClient       func()
	MaxReadDuration   *time.Duration
}

func init() {
	maxDuration, _ = time.ParseDuration("10m")
	defaultRetryInterval, _ = time.ParseDuration("10s")
}

func (f *FluentdFunctionalFramework) GetMaxReadDuration() time.Duration {
	if f.MaxReadDuration != nil {
		return *f.MaxReadDuration
	}
	return maxDuration
}

func NewFluentdFunctionalFramework() *FluentdFunctionalFramework {
	test := client.NewTest()
	return NewFluentdFunctionalFrameworkUsing(test, test.Close, 0)
}

func NewFluentdFunctionalFrameworkForTest(t *testing.T) *FluentdFunctionalFramework {
	return NewFluentdFunctionalFrameworkUsing(client.ForTest(t), func() {}, 0)
}

func NewFluentdFunctionalFrameworkUsing(t *client.Test, fnClose func(), verbosity int) *FluentdFunctionalFramework {
	if level, found := os.LookupEnv("LOG_LEVEL"); found {
		if i, err := strconv.Atoi(level); err == nil {
			verbosity = i
		}
	}

	log.MustInit("functional-framework")
	log.SetLogLevel(verbosity)
	testName := "functional"
	framework := &FluentdFunctionalFramework{
		Name:      testName,
		Namespace: t.NS.Name,
		image:     utils.GetComponentImage(constants.FluentdName),
		labels: map[string]string{
			"testtype": "functional",
			"testname": testName,
		},
		Test:             t,
		Forwarder:        runtime.NewClusterLogForwarder(),
		receiverBuilders: []receiverBuilder{},
		closeClient:      fnClose,
	}
	return framework
}

func NewPartialCRIOLogMessage(timestamp, message string) string {
	return NewCRIOLogMessage(timestamp, message, true)
}

func NewFullCRIOLogMessage(timestamp, message string) string {
	return NewCRIOLogMessage(timestamp, message, false)
}

func NewCRIOLogMessage(timestamp, message string, partial bool) string {
	fullOrPartial := "F"
	if partial {
		fullOrPartial = "P"
	}
	return fmt.Sprintf("%s stdout %s %s", timestamp, fullOrPartial, message)
}

// CRIOTime returns the CRIO string format of time t.
func CRIOTime(t time.Time) string { return time.Now().UTC().Format(time.RFC3339Nano) }

func (f *FluentdFunctionalFramework) Cleanup() {
	f.closeClient()
}

func (f *FluentdFunctionalFramework) RunCommand(container string, cmd ...string) (string, error) {
	log.V(2).Info("Running", "container", container, "cmd", cmd)
	out, err := runtime.ExecOc(f.Pod, strings.ToLower(container), cmd[0], cmd[1:]...)
	log.V(2).Info("Exec'd", "out", out, "err", err)
	return out, err
}

func (f *FluentdFunctionalFramework) AddOutputContainersVisitors() []runtime.PodBuilderVisitor {
	visitors := []runtime.PodBuilderVisitor{
		func(b *runtime.PodBuilder) error {
			return f.addOutputContainers(b, f.Forwarder.Spec.Outputs)
		},
	}
	return visitors
}

//Deploy the objects needed to functional Test
func (f *FluentdFunctionalFramework) Deploy() (err error) {
	return f.DeployWithVisitors(f.AddOutputContainersVisitors())
}

func (f *FluentdFunctionalFramework) DeployWithVisitor(visitor runtime.PodBuilderVisitor) (err error) {
	visitors := []runtime.PodBuilderVisitor{
		visitor,
	}
	return f.DeployWithVisitors(visitors)
}

//Deploy the objects needed to functional Test
func (f *FluentdFunctionalFramework) DeployWithVisitors(visitors []runtime.PodBuilderVisitor) (err error) {
	log.V(2).Info("Generating config", "forwarder", f.Forwarder)
	clfYaml, _ := yaml.Marshal(f.Forwarder)
	debugOutput := false
	if f.Conf, err = forwarder.Generate(string(clfYaml), false, debugOutput); err != nil {
		return err
	}
	//remove journal for now since we can not mimic it
	re := regexp.MustCompile(`(?msU).*<source>.*type.*systemd.*</source>`)
	f.Conf = string(re.ReplaceAll([]byte(f.Conf), []byte{}))
	log.V(2).Info("Generating Certificates")
	if err, _, _ = certificates.GenerateCertificates(f.Test.NS.Name,
		test.GitRoot("scripts"), "elasticsearch",
		utils.DefaultWorkingDir); err != nil {
		return err
	}
	log.V(2).Info("Creating config configmap")
	configmap := runtime.NewConfigMap(f.Test.NS.Name, f.Name, map[string]string{})

	//create dirs that dont exist in testing
	replace := `#!/bin/bash
mkdir -p /var/log/{kube-apiserver,oauth-apiserver,audit,ovn}
for d in kube-apiserver oauth-apiserver audit; do
  touch /var/log/$d/{audit.log,acl-audit-log.log}
done
`
	runScript := strings.Replace(fluentd.RunScript, "#!/bin/bash\n", replace, 1)
	runtime.NewConfigMapBuilder(configmap).
		Add("fluent.conf", f.Conf).
		Add("run.sh", runScript).
		Add("clfyaml", string(clfYaml))
	if err = f.Test.Client.Create(configmap); err != nil {
		return err
	}

	log.V(2).Info("Creating certs configmap")
	certsName := "certs-" + f.Name
	certs := runtime.NewConfigMap(f.Test.NS.Name, certsName, map[string]string{})
	runtime.NewConfigMapBuilder(certs).
		Add("tls.key", string(utils.GetWorkingDirFileContents("system.logging.fluentd.key"))).
		Add("tls.crt", string(utils.GetWorkingDirFileContents("system.logging.fluentd.crt")))
	if err = f.Test.Client.Create(certs); err != nil {
		return err
	}

	log.V(2).Info("Creating service")
	service := runtime.NewService(f.Test.NS.Name, f.Name)
	runtime.NewServiceBuilder(service).
		AddServicePort(24231, 24231).
		WithSelector(f.labels)
	if err = f.Test.Client.Create(service); err != nil {
		return err
	}

	role := runtime.NewRole(f.Test.NS.Name, f.Name,
		v1.PolicyRule{
			Verbs:     []string{"list", "get"},
			Resources: []string{"pods", "namespaces"},
			APIGroups: []string{""},
		},
	)
	if err = f.Test.Client.Create(role); err != nil {
		return err
	}
	rolebinding := runtime.NewRoleBinding(f.Test.NS.Name, f.Name,
		v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		},
		v1.Subject{
			Kind: "ServiceAccount",
			Name: "default",
		},
	)
	if err = f.Test.Client.Create(rolebinding); err != nil {
		return err
	}

	log.V(2).Info("Defining pod...")
	f.Pod = runtime.NewPod(f.Test.NS.Name, f.Name)
	b := runtime.NewPodBuilder(f.Pod).
		WithLabels(f.labels).
		AddConfigMapVolume("config", f.Name).
		AddConfigMapVolume("entrypoint", f.Name).
		AddConfigMapVolume("certs", certsName).
		AddContainer(constants.FluentdName, f.image).
		AddEnvVar("LOG_LEVEL", "debug").
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVar("NODE_NAME", FunctionalNodeName).
		AddVolumeMount("config", "/etc/fluent/configs.d/user", "", true).
		AddVolumeMount("entrypoint", "/opt/app-root/src/run.sh", "run.sh", true).
		AddVolumeMount("certs", "/etc/fluent/metrics", "", true).
		End()
	for _, visit := range visitors {
		if err = visit(b); err != nil {
			return err
		}
	}
	log.V(2).Info("Creating pod", "pod", f.Pod)
	if err = f.Test.Client.Create(f.Pod); err != nil {
		return err
	}

	log.V(2).Info("waiting for pod to be ready")
	if err = oc.Literal().From("oc wait -n %s pod/%s --timeout=120s --for=condition=Ready", f.Test.NS.Name, f.Name).Output(); err != nil {
		return err
	}
	if err = f.Test.Client.Get(f.Pod); err != nil {
		return err
	}
	log.V(2).Info("waiting for service endpoints to be ready")
	err = wait.PollImmediate(time.Second*2, time.Second*10, func() (bool, error) {
		ips, err := oc.Get().WithNamespace(f.Test.NS.Name).Resource("endpoints", f.Name).OutputJsonpath("{.subsets[*].addresses[*].ip}").Run()
		if err != nil {
			return false, nil
		}
		// if there are IPs in the service endpoint, the service is available
		if strings.TrimSpace(ips) != "" {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("service could not be started")
	}
	log.V(2).Info("waiting for fluentd to be ready")
	err = wait.PollImmediate(time.Second*2, time.Second*30, func() (bool, error) {
		output, err := oc.Literal().From("oc logs -n %s pod/%s %s", f.Test.NS.Name, f.Name, constants.FluentdName).Run()
		if err != nil {
			return false, nil
		}

		// if fluentd started successfully return success
		if strings.Contains(output, "flush_thread actually running") || strings.Contains(output, "fluentd worker is now running worker=0") || debugOutput {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("fluentd did not start in the container")
	}
	for _, cs := range f.Pod.Status.ContainerStatuses {
		if cs.Name == constants.FluentdName {
			f.fluentContainerId = strings.TrimPrefix(cs.ContainerID, "cri-o://")
			break
		}
	}
	return nil
}

func (f *FluentdFunctionalFramework) addOutputContainers(b *runtime.PodBuilder, outputs []logging.OutputSpec) error {
	log.V(2).Info("Adding outputs", "outputs", outputs)
	for _, output := range outputs {
		switch output.Type {
		case logging.OutputTypeFluentdForward:
			if err := f.AddForwardOutput(b, output); err != nil {
				return err
			}
		case logging.OutputTypeSyslog:
			if err := f.addSyslogOutput(b, output); err != nil {
				return err
			}
		case logging.OutputTypeElasticsearch:
			if err := f.addES7Output(b, output); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FluentdFunctionalFramework) WaitForPodToBeReady() error {
	return oc.Literal().From("oc wait -n %s pod/%s --timeout=60s --for=condition=Ready", f.Test.NS.Name, f.Name).Output()
}

func CreateAppLogFromJson(jsonstr string) string {
	jsonMsg := strings.ReplaceAll(jsonstr, "\n", "")
	timestamp := "2020-11-04T18:13:59.061892+00:00"

	return fmt.Sprintf("%s stdout F %s", timestamp, jsonMsg)
}

func (f *FluentdFunctionalFramework) WriteMessagesToApplicationLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s-%s.log", fluentdLogPath[applicationLog], f.Pod.Name, f.Pod.Namespace, constants.FluentdName, f.fluentContainerId)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *FluentdFunctionalFramework) WriteMessagesInfraContainerLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s-%s.log", fluentdLogPath[applicationLog], f.Pod.Name, "openshift-fake-infra", constants.FluentdName, f.fluentContainerId)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *FluentdFunctionalFramework) WriteMessagesToAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[auditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}
func (f *FluentdFunctionalFramework) WriteAuditHostLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[auditLog])
	msg := NewAuditHostLog(time.Now())
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *FluentdFunctionalFramework) WriteMessagesTok8sAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[k8sAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}
func (f *FluentdFunctionalFramework) WriteK8sAuditLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[k8sAuditLog])
	for numOfLogs > 0 {
		entry := NewKubeAuditLog(time.Now())
		if err := f.WriteMessagesToLog(entry, 1, filename); err != nil {
			return err
		}
		numOfLogs -= 1
	}
	return nil
}
func (f *FluentdFunctionalFramework) WriteOpenshiftAuditLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[OpenshiftAuditLog])
	for numOfLogs > 0 {
		now := CRIOTime(time.Now())
		entry := fmt.Sprintf(OpenShiftAuditLogTemplate, now, now)
		if err := f.WriteMessagesToLog(entry, 1, filename); err != nil {
			return err
		}
		numOfLogs -= 1
	}
	return nil
}

func (f *FluentdFunctionalFramework) WriteMessagesToOpenshiftAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[OpenshiftAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *FluentdFunctionalFramework) WriteMessagesToOVNAuditLog(msg string, numOfLogs int) error {

	filename := fmt.Sprintf("%s/acl-audit-log.log", fluentdLogPath[ovnAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}
func (f *FluentdFunctionalFramework) WriteOVNAuditLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/acl-audit-log.log", fluentdLogPath[ovnAuditLog])
	for numOfLogs > 0 {
		entry := NewOVNAuditLog(time.Now())
		if err := f.WriteMessagesToLog(entry, 1, filename); err != nil {
			return err
		}
		numOfLogs -= 1
	}
	return nil
}

func (f *FluentdFunctionalFramework) WritesApplicationLogs(numOfLogs uint64) error {
	return f.WritesNApplicationLogsOfSize(numOfLogs, uint64(100))
}

func (f *FluentdFunctionalFramework) WritesNApplicationLogsOfSize(numOfLogs, size uint64) error {
	msg := "$(date -u +'%Y-%m-%dT%H:%M:%S.%N%:z') stdout F $msg "
	//podname_ns_containername-containerid.log
	//functional_testhack-16511862744968_fluentd-90a0f0a7578d254eec640f08dd155cc2184178e793d0289dff4e7772757bb4f8.log
	filepath := fmt.Sprintf("/var/log/containers/%s_%s_%s-%s.log", f.Pod.Name, f.Pod.Namespace, constants.FluentdName, f.fluentContainerId)
	result, err := f.RunCommand(constants.FluentdName, "bash", "-c", fmt.Sprintf("mkdir -p /var/log/containers;echo > %s;msg=$(cat /dev/urandom|tr -dc 'a-zA-Z0-9'|fold -w %d|head -n 1);for n in $(seq 1 %d);do echo %s >> %s; done", filepath, size, numOfLogs, msg, filepath))
	log.V(3).Info("FluentdFunctionalFramework.WritesNApplicationLogsOfSize", "result", result, "err", err)
	return err
}

func (f *FluentdFunctionalFramework) WriteMessagesToLog(msg string, numOfLogs int, filename string) error {
	logPath := filepath.Dir(filename)
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	cmd := fmt.Sprintf("mkdir -p %s;for n in {1..%d};do echo \"$(echo %s|base64 -d)\" >> %s;sleep 1s;done", logPath, numOfLogs, encoded, filename)
	log.V(3).Info("Writing messages to log with command", "cmd", cmd)
	result, err := f.RunCommand(constants.FluentdName, "bash", "-c", cmd)
	log.V(3).Info("FluentdFunctionalFramework.WriteMessagesToLog", "result", result, "err", err)
	return err
}

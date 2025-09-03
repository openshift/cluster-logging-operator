package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	o "github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/openshift/origin/test/extended/util/compat_otp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

type Clusterlogforwarder struct {
	CollectApplicationLogs    bool // optional, if true, will add cluster-role/collect-application-logs to the serviceAccount
	CollectAuditLogs          bool // optional, if true, will add cluster-role/collect-audit-logs to the serviceAccount
	CollectInfrastructureLogs bool // optional, if true, will add cluster-role/collect-infrastructure-logs to the serviceAccount
	EnableMonitoring          bool // optional, if true, will add label `openshift.io/cluster-monitoring: "true"` to the project, and create role/prometheus-k8s rolebinding/prometheus-k8s in the namespace, works when when !(clf.namespace == "openshift-operators-redhat" || clf.namespace == "openshift-logging")
	Name                      string
	Namespace                 string
	ServiceAccountName        string
	TemplateFile              string // the template used to create clusterlogforwarder, no default value
	SecretName                string // optional, if it's specified, when creating CLF, the parameter `"SECRET_NAME="+clf.secretName` will be added automatically
	WaitForPodReady           bool   // optional, if true, will check daemonset stats
}

// create clusterlogforwarder CR from a template
func (clf *Clusterlogforwarder) Create(oc *exutil.CLI, optionalParameters ...string) {
	if clf.Namespace != "openshift-logging" {
		compat_otp.SetNamespacePrivileged(oc, clf.Namespace)
	}

	//parameters := []string{"-f", clf.templateFile, "--ignore-unknown-parameters=true", "-p", "NAME=" + clf.name, "NAMESPACE=" + clf.namespace}
	parameters := []string{"-f", clf.TemplateFile, "-p", "NAME=" + clf.Name, "NAMESPACE=" + clf.Namespace}
	if clf.SecretName != "" {
		parameters = append(parameters, "SECRET_NAME="+clf.SecretName)
	}

	clf.CreateServiceAccount(oc)
	parameters = append(parameters, "SERVICE_ACCOUNT_NAME="+clf.ServiceAccountName)

	if len(optionalParameters) > 0 {
		parameters = append(parameters, optionalParameters...)
	}

	file, processErr := ProcessTemplate(oc, parameters...)
	defer os.Remove(file)
	if processErr != nil {
		e2e.Failf("error processing file: %v", processErr)
	}
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", file, "-n", clf.Namespace).Execute()
	if err != nil {
		e2e.Failf("error creating clusterlogforwarder: %v", err)
	}
	o.Expect(WaitForResourceToAppear(oc, "clusterlogforwarders.observability.openshift.io", clf.Name, clf.Namespace)).NotTo(o.HaveOccurred())
	if clf.WaitForPodReady {
		clf.WaitForCollectorPodsReady(oc)
	}

	if clf.Namespace != CloNS && clf.Namespace != LoNS && clf.EnableMonitoring {
		EnableClusterMonitoring(oc, clf.Namespace)
	}
}

// CreateServiceAccount creates the serviceaccount and add the required clusterroles to the serviceaccount
func (clf *Clusterlogforwarder) CreateServiceAccount(oc *exutil.CLI) {
	_, err := oc.AdminKubeClient().CoreV1().ServiceAccounts(clf.Namespace).Get(context.Background(), clf.ServiceAccountName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		err = CreateServiceAccount(oc, clf.Namespace, clf.ServiceAccountName)
		if err != nil {
			e2e.Failf("can't create the serviceaccount: %v", err)
		}
	}
	if clf.CollectApplicationLogs {
		err = AddClusterRoleToServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "collect-application-logs")
		o.Expect(err).NotTo(o.HaveOccurred())
	}
	if clf.CollectInfrastructureLogs {
		err = AddClusterRoleToServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "collect-infrastructure-logs")
		o.Expect(err).NotTo(o.HaveOccurred())
	}
	if clf.CollectAuditLogs {
		err = AddClusterRoleToServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "collect-audit-logs")
		o.Expect(err).NotTo(o.HaveOccurred())
	}
}

// update existing clusterlogforwarder CR
// if template is specified, then run command `oc process -f template -p patches | oc apply -f -`
// if template is not specified, then run command `oc patch clusterlogforwarder/${clf.name} -p patches`
// if use patch, should add `--type=` in the end of patches
func (clf *Clusterlogforwarder) Update(oc *exutil.CLI, template string, patches ...string) {
	var err error
	if template != "" {
		//parameters := []string{"-f", template, "--ignore-unknown-parameters=true", "-p", "NAME=" + clf.name, "NAMESPACE=" + clf.namespace}
		parameters := []string{"-f", template, "-p", "NAME=" + clf.Name, "NAMESPACE=" + clf.Namespace}
		if clf.SecretName != "" {
			parameters = append(parameters, "SECRET_NAME="+clf.SecretName)
		}
		parameters = append(parameters, "SERVICE_ACCOUNT_NAME="+clf.ServiceAccountName)

		if len(patches) > 0 {
			parameters = append(parameters, patches...)
		}
		file, processErr := ProcessTemplate(oc, parameters...)
		defer os.Remove(file)
		if processErr != nil {
			e2e.Failf("error processing file: %v", processErr)
		}
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", file, "-n", clf.Namespace).Execute()
	} else {
		parameters := []string{"clusterlogforwarders.observability.openshift.io/" + clf.Name, "-n", clf.Namespace, "-p"}
		parameters = append(parameters, patches...)
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args(parameters...).Execute()
	}
	if err != nil {
		e2e.Failf("error updating clusterlogforwarder: %v", err)
	}
}

// patch existing clusterlogforwarder CR and return the output,
// return patch_output and error
func (clf *Clusterlogforwarder) Patch(oc *exutil.CLI, patch_string string) (string, error) {
	parameters := []string{"clusterlogforwarders.observability.openshift.io/" + clf.Name, "-n", clf.Namespace, "-p"}
	parameters = append(parameters, patch_string, "--type=json")
	return oc.AsAdmin().WithoutNamespace().Run("patch").Args(parameters...).Output()
}

// delete the clusterlogforwarder CR
func (clf *Clusterlogforwarder) Delete(oc *exutil.CLI) {
	o.Expect(DeleteResourceFromCluster(oc, "clusterlogforwarders.observability.openshift.io", clf.Name, clf.Namespace)).NotTo(o.HaveOccurred())

	if clf.CollectApplicationLogs {
		RemoveClusterRoleFromServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "collect-application-logs")
	}
	if clf.CollectInfrastructureLogs {
		RemoveClusterRoleFromServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "collect-infrastructure-logs")
	}
	if clf.CollectAuditLogs {
		RemoveClusterRoleFromServiceAccount(oc, clf.Namespace, clf.ServiceAccountName, "collect-audit-logs")
	}
	o.Expect(DeleteResourceFromCluster(oc, "serviceaccount", clf.Name, clf.Namespace)).NotTo(o.HaveOccurred())
	o.Expect(WaitUntilResourceIsGone(oc, "daemonset", clf.Name, clf.Namespace)).NotTo(o.HaveOccurred())
}

func (clf *Clusterlogforwarder) WaitForCollectorPodsReady(oc *exutil.CLI) {
	WaitForDaemonsetPodsToBeReady(oc, clf.Namespace, clf.Name)
}

func (clf *Clusterlogforwarder) GetCollectorNodeNames(oc *exutil.CLI) ([]string, error) {
	var nodes []string
	pods, err := oc.AdminKubeClient().CoreV1().Pods(clf.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=collector,app.kubernetes.io/instance=" + clf.Name})
	for _, pod := range pods.Items {
		nodes = append(nodes, pod.Spec.NodeName)
	}
	return nodes, err

}

type LogFileMetricExporter struct {
	Name          string
	Namespace     string
	Template      string
	WaitPodsReady bool
}

func (lfme *LogFileMetricExporter) Create(oc *exutil.CLI, optionalParameters ...string) {
	if lfme.Name == "" {
		lfme.Name = "instance"
	}
	if lfme.Namespace == "" {
		lfme.Namespace = LoggingNS
	}
	if lfme.Template == "" {
		lfme.Template = compat_otp.FixturePath("testdata", "logging", "logfilemetricexporter", "lfme.yaml")
	}

	parameters := []string{"-f", lfme.Template, "-p", "NAME=" + lfme.Name, "NAMESPACE=" + lfme.Namespace}
	if len(optionalParameters) > 0 {
		parameters = append(parameters, optionalParameters...)
	}

	file, processErr := ProcessTemplate(oc, parameters...)
	defer os.Remove(file)
	if processErr != nil {
		e2e.Failf("error processing file: %v", processErr)
	}
	err := oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", file, "-n", lfme.Namespace).Execute()
	if err != nil {
		e2e.Failf("error creating logfilemetricexporter: %v", err)
	}
	o.Expect(WaitForResourceToAppear(oc, "logfilemetricexporter", lfme.Name, lfme.Namespace)).NotTo(o.HaveOccurred())
	if lfme.WaitPodsReady {
		WaitForDaemonsetPodsToBeReady(oc, lfme.Namespace, "logfilesmetricexporter")
	}
}

func (lfme *LogFileMetricExporter) Delete(oc *exutil.CLI) {
	o.Expect(DeleteResourceFromCluster(oc, "logfilemetricexporter", lfme.Name, lfme.Namespace)).NotTo(o.HaveOccurred())
	o.Expect(WaitUntilResourceIsGone(oc, "daemonset", "logfilesmetricexporter", lfme.Namespace)).NotTo(o.HaveOccurred())
}

func CheckCollectorConfiguration(oc *exutil.CLI, ns, cmName string, searchStrings ...string) (bool, error) {
	// Parse the vector.toml file
	dirname := "/tmp/" + oc.Namespace() + "-vectortoml"
	defer os.RemoveAll(dirname)
	err := os.MkdirAll(dirname, 0777)
	if err != nil {
		return false, err
	}

	_, err = oc.AsAdmin().WithoutNamespace().Run("extract").Args("configmap/"+cmName, "-n", ns, "--confirm", "--to="+dirname).Output()
	if err != nil {
		return false, err
	}

	filename := filepath.Join(dirname, "vector.toml")
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	for _, s := range searchStrings {
		if !strings.Contains(string(content), s) {
			e2e.Logf("can't find %s in vector.toml", s)
			return false, nil
		}
	}
	return true, nil
}

func RunLoggingMustGather(oc *exutil.CLI) ([]string, error) {
	// create a temporary directory
	baseDir := compat_otp.FixturePath("testdata", "logging")
	mustGatherPath := filepath.Join(baseDir, "temp"+GetRandomString())
	defer exec.Command("rm", "-r", mustGatherPath).Output()
	err := os.MkdirAll(mustGatherPath, 0755)
	o.Expect(err).NotTo(o.HaveOccurred())

	image, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("deployment.apps/cluster-logging-operator", `-ojsonpath={.spec.template.spec.containers[?(@.name == "cluster-logging-operator")].image}`, "-n", CloNS).Output()
	if err != nil {
		return []string{}, fmt.Errorf("can't get CLO image: %s", err)
	}

	err = oc.AsAdmin().WithoutNamespace().Run("adm").Args("must-gather", "--image="+image, "--dest-dir="+mustGatherPath).Execute()
	if err != nil {
		return []string{}, fmt.Errorf("can't get logging must-gather: %v", err)
	}

	cloImgID, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", CloNS, "pods", "-l", "name=cluster-logging-operator", "-o", "jsonpath={.items[0].status.containerStatuses[0].imageID}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	replacer := strings.NewReplacer(".", "-", "/", "-", ":", "-", "@", "-")
	cloImgDir := replacer.Replace(cloImgID)
	return ListFilesAndDirectories(mustGatherPath + "/" + cloImgDir)
}

func CheckCLOMetric(oc *exutil.CLI, token string) {
	//metrics exposed by cluster-logging-operator
	metrics := []string{"openshift_logging:log_forwarder_pipelines:sum", "openshift_logging:log_forwarders:sum", "openshift_logging:log_forwarder_input_type:sum", "openshift_logging:log_forwarder_output_type:sum", "container_file_descriptors", "kubelet_running_containers", "container_cpu_usage_seconds_total"}
	//merics need more time
	slowMetrics := []string{"cluster:log_logged_bytes_total:sum", "collector:received_events:sum_rate", "openshift_logging:vector_component_received_bytes_total:rate5m", "vector_buffer_byte_size{buffer_type=\"disk\"}", "vector_component_sent_bytes_total", "vector_open_files", "vector_component_received_event_bytes_total"}
	var missMetics []string
	for _, metricName := range metrics {
		if !FindMetric(oc, token, metricName, 3) {
			missMetics = append(missMetics, metricName)
		}
	}
	for _, metricName := range slowMetrics {
		if !FindMetric(oc, token, metricName, 5) {
			missMetics = append(missMetics, metricName)
		}
	}
	o.Expect(len(missMetics) == 0).Should(o.BeTrue(), fmt.Sprintf("can't find metrics %v in given time", missMetics))
}

type eventRouter struct {
	name      string
	namespace string
	template  string
}

func (e eventRouter) deploy(oc *exutil.CLI, optionalParameters ...string) {
	parameters := []string{"-f", e.template, "-l", "app=eventrouter", "-p", "NAME=" + e.name, "NAMESPACE=" + e.namespace}
	if len(optionalParameters) > 0 {
		parameters = append(parameters, optionalParameters...)
	}

	file, processErr := ProcessTemplate(oc, parameters...)
	defer os.Remove(file)
	if processErr != nil {
		e2e.Failf("error processing file: %v", processErr)
	}
	err := oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", file, "-n", e.namespace).Execute()
	if err != nil {
		e2e.Failf("error deploying eventrouter: %v", err)
	}
	o.Expect(WaitForResourceToAppear(oc, "deployment", e.name, e.namespace)).NotTo(o.HaveOccurred())
	WaitForDeploymentPodsToBeReady(oc, e.namespace, e.name)
}

func (e eventRouter) delete(oc *exutil.CLI) {
	_ = DeleteResourceFromCluster(oc, "deployment", e.name, e.namespace)
	_ = DeleteResourceFromCluster(oc, "configmaps", e.name, e.namespace)
	_ = DeleteResourceFromCluster(oc, "serviceaccounts", e.name, e.namespace)
	_ = DeleteResourceFromCluster(oc, "clusterrolebindings", e.name+"-reader-binding", "")
	_ = DeleteResourceFromCluster(oc, "clusterrole", e.name+"-reader", "")
}

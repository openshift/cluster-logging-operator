package normalization

import (
	"fmt"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

const (
	deletedPodNamespace = "deleted-ns"
	deletedPodName      = "deleted-pod"
	deletedPodUID       = "aabbccdd-1234-5678-abcd-ef0123456789"
	deletedContainer    = "deleted-container"
)

var deletedPodLogPath = fmt.Sprintf("/var/log/pods/%s_%s_%s/%s/0.log",
	deletedPodNamespace, deletedPodName, deletedPodUID, deletedContainer)

var fakeContainerSource = fmt.Sprintf(`
[sources.input_application_container]
type = "file"
include = ["%s"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

`, deletedPodLogPath)

var _ = Describe("[Functional][Normalization] container metadata fallback from file path (LOG-9584)", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToElasticSearchOutput()

		framework.VisitConfig = func(conf string) string {
			// Match the entire kubernetes_logs source block including sub-tables
			// (pod_annotation_fields, namespace_annotation_fields)
			re := regexp.MustCompile(`(?ms)\[sources\.input_application_container\].*?\[sources\.input_application_container\.namespace_annotation_fields\][^\[]*`)
			return string(re.ReplaceAll([]byte(conf), []byte(fakeContainerSource)))
		}
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should extract kubernetes metadata from file path when pod metadata annotation is missing", func() {
		msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "log from deleted pod", false)
		Expect(framework.WriteMessagesToLog(msg, 1, deletedPodLogPath)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(logs).ToNot(BeEmpty())

		log := logs[0]
		Expect(log.Kubernetes.NamespaceName).To(Equal(deletedPodNamespace))
		Expect(log.Kubernetes.PodName).To(Equal(deletedPodName))
		Expect(log.Kubernetes.PodID).To(Equal(deletedPodUID))
		Expect(log.Kubernetes.ContainerName).To(Equal(deletedContainer))
		Expect(log.LogType).To(Equal("application"))
	})
})

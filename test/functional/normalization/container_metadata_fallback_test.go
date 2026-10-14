package normalization

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
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
			// Match the entire kubernetes_logs source block including sub-tables (pod_annotation_fields, namespace_annotation_fields)
			// Find the start of the input_application_container section
			startIdx := strings.Index(conf, "[sources.input_application_container]")
			if startIdx == -1 {
				return conf
			}
			// Find the next section marker (starts with '[') after the input_application_container block
			re := regexp.MustCompile(`\n\[`)
			nextSectionMatch := re.FindStringIndex(conf[startIdx+1:])
			if nextSectionMatch == nil {
				// No next section, replace until end of string
				return conf[:startIdx] + fakeContainerSource
			}
			endIdx := startIdx + 1 + nextSectionMatch[0]
			return conf[:startIdx] + fakeContainerSource + conf[endIdx:]
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

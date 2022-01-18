package normalization

import (
	"fmt"
	functional "github.com/openshift/cluster-logging-operator/test/framework/functional"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

var _ = Describe("[Functional][Normalization] Fluentd normalization", func() {

	var (
		framework *functional.CollectorFunctionalFramework
		pb        *functional.PipelineBuilder
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		pb = functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication)
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("for ES output", func() {
		It("should remove 'kubernetes.labels' and create 'kubernetes.flat_labels' with an array of 'kubernetes.labels'", func() {
			pb.ToElasticSearchOutput()
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogs(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			//verify the new key exists
			Expect(logs[0].Kubernetes.FlatLabels).To(Not(BeNil()), fmt.Sprintf("Expected to find the kubernetes.flat_labels key in %#v", logs[0]))

			//verify we removed the old key
			Expect(logs[0].Kubernetes.Labels).To(BeNil(), fmt.Sprintf("Did not expect to find the kubernetes.labels key in %#v", logs[0]))
		})
	})
	Context("for non-ES output", func() {
		It("should not remove 'kubernetes.labels' and not add 'kubernetes.flat_labels'", func() {
			pb.ToFluentForwardOutput()
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeFluentdForward)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			//verify the new key does not exists
			Expect(logs[0].Kubernetes.FlatLabels).To(BeNil(), fmt.Sprintf("Did not expect to find the kubernetes.flat_labels key in %#v", logs[0]))

			//verify the old key exists
			Expect(logs[0].Kubernetes.Labels).To(Not(BeNil()), fmt.Sprintf("Expected to find the kubernetes.labels key in %#v", logs[0]))
		})
	})

})

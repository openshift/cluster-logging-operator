package multiple

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var _ = Describe("[Functional][Outputs][Multiple]", func() {

	var (
		framework *functional.CollectorFunctionalFramework
		builder   *functional.PipelineBuilder
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		builder = functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(loggingv1.InputNameApplication)
		builder.ToFluentForwardOutput()
		builder.ToElasticSearchOutput()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Describe("when multiple outputs are configured", func() {

		Describe("and both are accepting logs", func() {

			BeforeEach(func() {
				Expect(framework.Deploy()).To(BeNil())
				Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			})

			It("should send logs to the fluentd receiver and elasticsearch", func() {

				logs, err := framework.ReadApplicationLogsFrom(loggingv1.OutputTypeFluentdForward)
				Expect(err).To(BeNil(), "Expected no error reading logs from %s", loggingv1.OutputTypeFluentdForward)
				Expect(logs).To(HaveLen(1), "Exp. to receive a log message at output type %s", loggingv1.OutputTypeFluentdForward)

				raw, err := framework.GetLogsFromElasticSearch(loggingv1.OutputTypeElasticsearch, loggingv1.InputNameApplication)
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				Expect(raw).To(Not(BeEmpty()))
				err = types.StrictlyParseLogsFromSlice(raw, &logs)
				Expect(err).To(BeNil(), "Expected no errors parsing the logs")
				Expect(logs).To(HaveLen(1), "Exp. to receive a log message at output type %s", loggingv1.OutputTypeElasticsearch)

			})
		})

		Describe("and one store is not available", func() {
			BeforeEach(func() {
				Expect(framework.DeployWithVisitor(func(builder *runtime.PodBuilder) error {
					return framework.AddForwardOutput(builder, framework.Forwarder.Spec.Outputs[0])
				})).To(BeNil())
				Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			})
			It("should send logs to the fluentd receiver only", func() {
				logs, err := framework.ReadApplicationLogsFrom(loggingv1.OutputTypeFluentdForward)
				Expect(err).To(BeNil(), "Expected no error reading logs from %s", loggingv1.OutputTypeFluentdForward)
				Expect(logs).To(HaveLen(1), "Exp. to receive a log message at output type %s", loggingv1.OutputTypeFluentdForward)

			})
		})

	})
})

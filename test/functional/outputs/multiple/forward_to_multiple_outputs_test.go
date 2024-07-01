package multiple

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var _ = Describe("[Functional][Outputs][Multiple]", func() {

	var (
		framework *functional.CollectorFunctionalFramework
		builder   *testruntime.PipelineBuilder
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		builder = testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication)
		builder.ToHttpOutput()
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

			It("should send logs to the http receiver and elasticsearch", func() {
				logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
				Expect(err).To(BeNil(), "Expected no error reading logs from %s", obs.OutputTypeHTTP)
				Expect(logs).To(HaveLen(1), "Exp. to receive a log message at output type %s", obs.OutputTypeHTTP)

				raw, err := framework.GetLogsFromElasticSearch(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication))
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				Expect(raw).To(Not(BeEmpty()))
				err = types.StrictlyParseLogsFromSlice(raw, &logs)
				Expect(err).To(BeNil(), "Expected no errors parsing the logs")
				Expect(logs).To(HaveLen(1), "Exp. to receive a log message at output type %s", obs.OutputTypeElasticsearch)

			})
		})

		Describe("and one store is not available", func() {
			BeforeEach(func() {
				Expect(framework.DeployWithVisitor(func(builder *runtime.PodBuilder) error {
					return framework.AddFluentdHttpOutput(builder, framework.Forwarder.Spec.Outputs[1])
				})).To(BeNil())
				Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			})
			It("should send logs to the http receiver only", func() {
				logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
				Expect(err).To(BeNil(), "Expected no error reading logs from %s", obs.OutputTypeHTTP)
				Expect(logs).To(HaveLen(1), "Exp. to receive a log message at output type %s", obs.OutputTypeHTTP)
			})
		})
	})
})

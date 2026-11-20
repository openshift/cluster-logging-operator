package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Validating output template values (LOG-10361)", func() {

	// The collector config is serialized as YAML (quoted/escaped or block
	// scalar), so values containing the former TOML literal multiline
	// terminator or newlines can no longer break out of the generated VRL.
	// Admission must accept them; safety holds by construction.
	It("should accept a kafka topic containing the former TOML literal terminator and a newline", func() {
		out := obs.OutputSpec{
			Name: "kafka-out",
			Type: obs.OutputTypeKafka,
			Kafka: &obs.Kafka{
				URL:   "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
				Topic: "top'''ic\nline2",
			},
		}
		context := internalcontext.ForwarderContext{
			Forwarder: &obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{out},
					Pipelines: []obs.PipelineSpec{
						{Name: "dummy", OutputRefs: internalobs.Outputs([]obs.OutputSpec{out}).Names()},
					},
				},
			},
			AdditionalContext: utils.Options{},
		}
		Validate(context)
		Expect(context.Forwarder.Status.OutputConditions).To(HaveLen(1))
		Expect(context.Forwarder.Status.OutputConditions[0]).To(matchers.MatchCondition("kafka-out", true, obs.ReasonValidationSuccess, ".*"))
	})
})

package otlp

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types/otlp"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"golang.org/x/exp/maps"
)

var _ = Describe("[Functional][Outputs][OTLP] Functional tests", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When forwarding logs to an otel-collector", func() {

		var (
			openShiftLabels = map[string]string{"foo": "bar"}
		)
		It("should send journal logs with correct resource attributes", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeInfrastructure).
				WithLabelsFilter(openShiftLabels).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			// Log message data
			logline := functional.NewJournalLog(3, "my journal message", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			// Read line from Log Forward output
			raw, err := framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeOTLP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())

			logs, err := otlp.ParseLogs(raw[0])
			resourceLog := logs.ResourceLogs[0]
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			expResourceAttributes := map[string]types.GomegaMatcher{
				otlp.OpenshiftClusterUID: MatchRegexp(".*"),
				otlp.OpenshiftLogSource:  BeEquivalentTo(obs.InfrastructureSourceNode),
				otlp.OpenshiftLogType:    BeEquivalentTo(obs.InputTypeInfrastructure),
				otlp.K8sNodeName:         MatchRegexp(".*"),
				otlp.ProcessExeName:      MatchRegexp(".*"),
				otlp.ProcessExePath:      MatchRegexp(".*"),
				otlp.ProcessCommandLine:  MatchRegexp(".*"),
				otlp.ProcessPID:          MatchRegexp(".*"),
				otlp.ServiceName:         MatchRegexp(".*"),
				// deprecated
				otlp.LogSource:          BeEquivalentTo(obs.InfrastructureSourceNode),
				otlp.LogType:            BeEquivalentTo(obs.InputTypeInfrastructure),
				otlp.KubernetesHost:     MatchRegexp(".*"),
				otlp.OpenshiftClusterID: MatchRegexp(".*"),
			}
			for key, value := range openShiftLabels {
				expResourceAttributes[fmt.Sprintf("%s%s", otlp.OpenshiftLabelPrefix, key)] = BeEquivalentTo(value)
			}

			resource := resourceLog.Resource
			Expect(otlp.CollectNames(resource.Attributes)).To(ContainElements(maps.Keys(expResourceAttributes)))
			for key, match := range expResourceAttributes {
				Expect(resource.Attribute(key).String()).To(match)
			}

			scopeLogs := resourceLog.ScopeLogs
			Expect(scopeLogs).To(HaveLen(1), "Expected a single scopeLog")

			logRecords := scopeLogs[0].LogRecords
			Expect(logRecords).To(HaveLen(1), "Expected same count as sent for log records")

			// Inspect the first log record for correct fields
			logRecord := logRecords[0]
			Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")
			Expect(logRecord.TimeUnixNano).To(MatchRegexp("[0-9]*"), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.ObservedTimeUnixNano).To(MatchRegexp("[0-9]*"))
			Expect(logRecord.SeverityText).To(Equal("err"))

			logAttributeNames := otlp.CollectNames(logRecord.Attributes)
			Expect(logAttributeNames).ToNot(ContainElement(HavePrefix(otlp.SystemdPrefix)), "Exp. no attributes with this prefix")
			// (Deprecated) compatibility attribute
			Expect(logAttributeNames).To(ContainElement(otlp.Level), "Expect logRecord attributes to contain level")
			Expect(logRecord.Attribute(otlp.Level).String()).To(Equal("err"))
		})
	})
})

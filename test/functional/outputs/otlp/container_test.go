package otlp

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types/otlp"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"golang.org/x/exp/maps"
)

var _ = Describe("[Functional][Outputs][OTLP] Functional tests", func() {
	const (
		timestamp     = "2023-08-28T12:59:28.573159188+00:00"
		timestampNano = "1693227568573159188"
	)

	var (
		framework    *functional.CollectorFunctionalFramework
		appNamespace string
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		framework.MaxReadDuration = utils.GetPtr(time.Second * 45)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When forwarding logs to an otel-collector", func() {
		var (
			openShiftLabels = map[string]string{"foo": "bar"}
		)
		DescribeTable("should send application logs with correct grouping and resource attributes", func(compression string) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithLabelsFilter(openShiftLabels).
				ToOtlpOutput(func(output *obs.OutputSpec) {
					output.OTLP.Tuning = &obs.OTLPTuningSpec{
						Compression: compression,
					}
				})

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			appNamespace = framework.Pod.Namespace

			// Write message to namespace
			messages := []string{"Format me to a container message!", "My second message.", "My third and final message."}
			for _, m := range messages {
				crioLine := functional.NewCRIOLogMessage(timestamp, m, false)
				Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())
			}

			// Read logs
			raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeOTLP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())

			logs, err := otlp.ParseLogs(raw[0])
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			// Access the first (and only) ResourceLog
			// Inspect resource attributes
			resourceLog := logs.ResourceLogs[0]
			resource := resourceLog.Resource

			expResourceAttributes := map[string]types.GomegaMatcher{
				otlp.OpenshiftClusterUID: MatchRegexp(".*"),
				otlp.OpenshiftLogSource:  BeEquivalentTo(obs.InfrastructureSourceContainer),
				otlp.OpenshiftLogType:    BeEquivalentTo(obs.InputTypeApplication),
				otlp.K8sNodeName:         MatchRegexp(".*"),
				otlp.K8sNamespaceName:    BeEquivalentTo(appNamespace),
				otlp.K8sContainerName:    MatchRegexp(".*"),
				otlp.K8sPodName:          BeEquivalentTo(framework.Pod.Name),
				otlp.K8sPodID:            BeEquivalentTo(framework.Pod.UID),
				// deprecated
				otlp.LogSource:               BeEquivalentTo(obs.InfrastructureSourceContainer),
				otlp.LogType:                 BeEquivalentTo(obs.InputTypeApplication),
				otlp.KubernetesContainerName: MatchRegexp(".*"),
				otlp.KubernetesHost:          MatchRegexp(".*"),
				otlp.KubernetesNamespaceName: BeEquivalentTo(appNamespace),
				otlp.KubernetesPodName:       BeEquivalentTo(framework.Pod.Name),
				otlp.OpenshiftClusterID:      MatchRegexp(".*"),
			}
			for key, value := range framework.Pod.Labels {
				expResourceAttributes[fmt.Sprintf("k8s.pod.label.%s", key)] = BeEquivalentTo(value)
			}
			for key, value := range openShiftLabels {
				expResourceAttributes[fmt.Sprintf("openshift.label.%s", key)] = BeEquivalentTo(value)
			}
			Expect(otlp.CollectNames(resource.Attributes)).To(ConsistOf(maps.Keys(expResourceAttributes)))
			for key, match := range expResourceAttributes {
				Expect(resource.Attribute(key).String()).To(match)
			}

			scopeLogs := resourceLog.ScopeLogs
			Expect(scopeLogs).To(HaveLen(1), "Expected a single scopeLog")
			logRecords := scopeLogs[0].LogRecords
			Expect(logRecords).To(HaveLen(1), "Expected same count group size defined by the transform")

			// Inspect the first log record for correct fields
			logRecord := logRecords[0]
			Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")
			Expect(logRecord.TimeUnixNano).To(Equal(timestampNano), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.ObservedTimeUnixNano).To(MatchRegexp("[0-9]*"))
			Expect(logRecord.SeverityText).To(Equal("default"))

			logAttributeNames := otlp.CollectNames(logRecord.Attributes)
			Expect(logAttributeNames).To(ContainElement(otlp.LogIOStream), "Expect logRecord to have 'log.iostream' attribute")
			Expect(logRecord.Attribute(otlp.LogIOStream).String()).To(Equal("stdout"))
			// (Deprecated) compatibility attribute
			Expect(logAttributeNames).To(ContainElement(otlp.Level), "Expect logRecord to have 'level' attribute")
			Expect(logRecord.Attribute(otlp.Level).String()).To(Equal("default"))
		},
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with zlib", "zlib"),
			Entry("should pass with zstd", "zstd"),
			Entry("should pass with no compression", "none"))
	})

	DescribeTable("should send logs with trace context", func(message string, expectedTraceID, expectedSpanID string, expectedFlag uint32) {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToOtlpOutput()

		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
		})).To(BeNil())

		appNamespace = framework.Pod.Namespace

		// Write message with trace context
		crioLine := functional.NewCRIOLogMessage(timestamp, message, false)
		Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())

		// Read logs
		raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeOTLP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(raw).ToNot(BeEmpty())

		logs, err := otlp.ParseLogs(raw[0])
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		resourceLog := logs.ResourceLogs[0]
		scopeLogs := resourceLog.ScopeLogs
		Expect(scopeLogs).To(HaveLen(1))
		logRecords := scopeLogs[0].LogRecords
		Expect(logRecords).To(HaveLen(1))

		logRecord := logRecords[0]
		Expect(logRecord.TraceID).To(Equal(expectedTraceID), "Expected trace_id to match")
		Expect(logRecord.SpanID).To(Equal(expectedSpanID), "Expected span_id to match")
		Expect(logRecord.Flags).To(Equal(expectedFlag), "Expected trace_flags to match")
	},
		Entry("should extract trace context with '=' separator",
			`Processing request trace_id="0af7651916cd43dd8448eb211c80319c" span_id="b7ad6b7169203331" trace_flags="01"`,
			"0af7651916cd43dd8448eb211c80319c", "b7ad6b7169203331", uint32(1),
		),
		Entry("should extract trace context with ':' separator",
			`host:192.168.0.1 trace_id:abcdef1234567890abcdef1234567890 span_id:fedcba9876543210 trace_flags:00 status:200`,
			"abcdef1234567890abcdef1234567890", "fedcba9876543210", uint32(0),
		),
		Entry("should convert uppercase hex to lowercase",
			`trace_id=ABCDEF1234567890ABCDEF1234567890 span_id=FEDCBA9876543210 trace_flags=00`,
			"abcdef1234567890abcdef1234567890", "fedcba9876543210", uint32(0),
		),
		Entry("should extract trace context without quotes",
			`trace_id=0af7651916cd43dd8448eb211c80319c span_id=b7ad6b7169203331 trace_flags=01`,
			"0af7651916cd43dd8448eb211c80319c", "b7ad6b7169203331", uint32(1),
		),
		Entry("should extract trace context with single quotes",
			`trace_id='0af7651916cd43dd8448eb211c80319c' span_id='b7ad6b7169203331' trace_flags='01'`,
			"0af7651916cd43dd8448eb211c80319c", "b7ad6b7169203331", uint32(1),
		),
		Entry("should extract trace context flag with single digit",
			`trace_id='0af7651916cd43dd8448eb211c80319c' span_id='b7ad6b7169203331' trace_flags='1'`,
			"0af7651916cd43dd8448eb211c80319c", "b7ad6b7169203331", uint32(1),
		),
	)
})

package otlp

import (
	"fmt"
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
	"time"
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
				otlp.NodeName:            MatchRegexp(".*"),
				otlp.K8sNamespaceName:    BeEquivalentTo(appNamespace),
				otlp.K8sContainerName:    MatchRegexp(".*"),
				otlp.K8sPodName:          BeEquivalentTo(framework.Pod.Name),
				otlp.K8sPodID:            BeEquivalentTo(framework.Pod.UID),
				// deprecated
				otlp.OpenshiftClusterID:      MatchRegexp(".*"),
				otlp.KubernetesHost:          MatchRegexp(".*"),
				otlp.LogSource:               BeEquivalentTo(obs.InfrastructureSourceContainer),
				otlp.LogType:                 BeEquivalentTo(obs.InputTypeApplication),
				otlp.KubernetesNamespaceName: BeEquivalentTo(appNamespace),
				otlp.KubernetesPodName:       BeEquivalentTo(framework.Pod.Name),
				otlp.KubernetesContainerName: MatchRegexp(".*"),
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
			Expect(logRecords).To(HaveLen(len(messages)), "Expected same count as sent for log records")

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
			Entry("should pass with no compression", "none"))
	})
})

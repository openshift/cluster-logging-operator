package otlp

import (
	"fmt"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	helpertypes "github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/helpers/types/otlp"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"golang.org/x/exp/maps"
	"k8s.io/apiserver/pkg/apis/audit"
)

var _ = Describe("[Functional][Outputs][OTLP] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		framework.MaxReadDuration = utils.GetPtr(time.Second * 45)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When forwarding Audit logs to an otel-collector", func() {
		var (
			openShiftLabels = map[string]string{"foo": "bar"}
		)

		DescribeTable("should send audit logs with correct grouping and resource attributes", func(auditSource obs.AuditSource,
			writeLog func(f *functional.CollectorFunctionalFramework, numberOfLogs int) (string, error),
			validateLogRecord func(logLine string, logRecord otlp.LogRecord)) {

			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				WithLabelsFilter(openShiftLabels).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			// Log message data
			totEntries := 3
			logLine, err := writeLog(framework, totEntries)
			Expect(err).To(BeNil())

			// Read line from Log Forward output
			raw, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeOTLP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())

			logs, err := otlp.ParseLogs(raw[0])
			resourceLog := logs.ResourceLogs[0]
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			resource := resourceLog.Resource

			expResourceAttributes := map[string]types.GomegaMatcher{
				otlp.OpenshiftClusterUID: MatchRegexp(".*"),
				otlp.OpenshiftLogSource:  BeEquivalentTo(auditSource),
				otlp.OpenshiftLogType:    BeEquivalentTo(obs.InputTypeAudit),
				otlp.K8sNodeName:         MatchRegexp(".*"),
				// deprecated
				otlp.LogSource:          BeEquivalentTo(auditSource),
				otlp.LogType:            BeEquivalentTo(obs.InputTypeAudit),
				otlp.KubernetesHost:     MatchRegexp(".*"),
				otlp.OpenshiftClusterID: MatchRegexp(".*"),
			}

			for key, value := range openShiftLabels {
				expResourceAttributes[fmt.Sprintf("%s%s", otlp.OpenshiftLabelPrefix, key)] = BeEquivalentTo(value)
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
			Expect(logRecord.TimeUnixNano).To(MatchRegexp("[0-9]*"), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.ObservedTimeUnixNano).To(MatchRegexp("[0-9]*"))

			validateLogRecord(logLine, logRecord)

		},
			Entry("with kubeAPI source", obs.AuditSourceKube, writeK8sAuditLog, validateK8sAPILogs),
			Entry("with openshiftAPI source", obs.AuditSourceOpenShift, writeOpenshiftAuditLog, validateOpenshiftAPILogs),
			Entry("with OVN source", obs.AuditSourceOVN, writeOvnAuditLog, validateOvnLog),
			Entry("with auditd (host) source", obs.AuditSourceAuditd, writeAuditHostLog, validateHostLog),
		)

	})
})

func writeK8sAuditLog(f *functional.CollectorFunctionalFramework, numOfLogs int) (string, error) {
	now := time.Now()
	logLine := functional.NewKubeAuditLog(now)
	return logLine, f.WriteMessagesTok8sAuditLog(logLine, numOfLogs)
}

func writeOpenshiftAuditLog(f *functional.CollectorFunctionalFramework, numOfLogs int) (string, error) {
	now := time.Now()
	nowCrio := functional.CRIOTime(now)
	logLine := fmt.Sprintf(functional.OpenShiftAuditLogTemplate, nowCrio, nowCrio)
	return logLine, f.WriteMessagesToOpenshiftAuditLog(logLine, numOfLogs)
}

func writeOvnAuditLog(f *functional.CollectorFunctionalFramework, numOfLogs int) (string, error) {
	now := time.Now()
	logLine := functional.NewOVNAuditLog(now)
	return logLine, f.WriteMessagesToOVNAuditLog(logLine, numOfLogs)
}

func writeAuditHostLog(f *functional.CollectorFunctionalFramework, numOfLogs int) (string, error) {
	now := time.Now()
	logLine := functional.NewAuditHostLog(now)
	return logLine, f.WriteMessagesToAuditLog(logLine, numOfLogs)
}

func validateHostLog(logLine string, logRecord otlp.LogRecord) {
	parts := strings.Split(logLine, "):")
	parts = strings.Split(parts[0], " ")
	Expect(len(parts)).To(BeNumerically(">", 1), "Not valid test data")
	auditdType := strings.Split(parts[0], "=")[1]
	sequence := strings.Split(parts[1], ":")[1]
	expLogAttributes := map[string]types.GomegaMatcher{
		otlp.AuditdType:  BeEquivalentTo(auditdType),
		otlp.LogSequence: BeEquivalentTo(sequence),
	}

	Expect(logRecord.SeverityText).To(BeEquivalentTo("default"))
	Expect(otlp.CollectNames(logRecord.Attributes)).To(ContainElements(maps.Keys(expLogAttributes)))
	expAttrs := maps.Keys(expLogAttributes)
	sort.Strings(expAttrs)
	for _, key := range expAttrs {
		Expect(logRecord.Attribute(key).String()).To(expLogAttributes[key], key)
	}
}

func validateOvnLog(logLine string, logRecord otlp.LogRecord) {
	parts := strings.Split(logLine, "|")
	Expect(len(parts)).To(BeNumerically(">", 3), "Not valid test data")
	sequence := parts[1]
	component := parts[2]
	severity := parts[3]

	Expect(logRecord.SeverityText).To(BeEquivalentTo(strings.ToLower(severity)))

	expLogAttributes := map[string]types.GomegaMatcher{
		otlp.LogSequence:     BeEquivalentTo(sequence),
		otlp.K8sOVNComponent: BeEquivalentTo(component),
	}
	Expect(otlp.CollectNames(logRecord.Attributes)).To(ContainElements(maps.Keys(expLogAttributes)))
	expAttrs := maps.Keys(expLogAttributes)
	sort.Strings(expAttrs)
	for _, key := range expAttrs {
		Expect(logRecord.Attribute(key).String()).To(expLogAttributes[key], key)
	}
}

func validateK8sAPILogs(logLine string, logRecord otlp.LogRecord) {
	event := &audit.Event{}
	test.MustUnmarshal(logLine, event)
	attributeNames := otlp.CollectNames(logRecord.Attributes)
	Expect(attributeNames).ToNot(ContainElement(HavePrefix("k8s.audit.")))
	Expect(attributeNames).ToNot(ContainElement("k8s.user."))
}

func validateOpenshiftAPILogs(logLine string, logRecord otlp.LogRecord) {
	event := &helpertypes.OpenshiftAuditLog{}
	test.MustUnmarshal(logLine, event)
	attributeNames := otlp.CollectNames(logRecord.Attributes)
	Expect(attributeNames).ToNot(ContainElement(HavePrefix("k8s.audit.")))
	Expect(attributeNames).ToNot(ContainElement("k8s.user."))
}

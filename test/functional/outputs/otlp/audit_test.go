package otlp

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
	"sort"
	"strconv"
	"strings"
	"time"
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
			writeLog func(f *functional.CollectorFunctionalFramework) (string, error),
			validateLogRecord func(logLine string, logRecord otlp.LogRecord)) {

			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				WithLabelsFilter(openShiftLabels).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			// Log message data
			logLine, err := writeLog(framework)
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
				otlp.OpenshiftLogType:    BeEquivalentTo(obs.InputTypeAudit),
				otlp.OpenshiftLogSource:  BeEquivalentTo(auditSource),
				otlp.NodeName:            MatchRegexp(".*"),
				// deprecated
				otlp.OpenshiftClusterID: MatchRegexp(".*"),
				otlp.KubernetesHost:     MatchRegexp(".*"),
				otlp.LogType:            BeEquivalentTo(obs.InputTypeAudit),
				otlp.LogSource:          BeEquivalentTo(auditSource),
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
			Expect(logRecords).To(HaveLen(3), "Expected same count as sent for log records")

			// Inspect the first log record for correct fields
			logRecord := logRecords[0]
			Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")
			Expect(logRecord.TimeUnixNano).To(MatchRegexp("[0-9]*"), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.ObservedTimeUnixNano).To(MatchRegexp("[0-9]*"))

			validateLogRecord(logLine, logRecord)

		},
			Entry("with kubeAPI source", obs.AuditSourceKube, writeK8sAuditLog, validateK8sAPILogs),
			//Entry("with openshiftAPI source", obs.AuditSourceOpenShift, writeOpenshiftAuditLog, validateOpenshiftAPILogs),
			//Entry("with OVN source", obs.AuditSourceOVN, writeOvnAuditLog, validateOvnLog),
			//Entry("with auditd (host) source", obs.AuditSourceAuditd, writeAuditHostLog, validateHostLog),
		)

	})
})

func writeK8sAuditLog(f *functional.CollectorFunctionalFramework) (string, error) {
	now := time.Now()
	logLine := functional.NewKubeAuditLog(now)
	return logLine, f.WriteMessagesTok8sAuditLog(logLine, 3)
}

func writeOpenshiftAuditLog(f *functional.CollectorFunctionalFramework) (string, error) {
	now := time.Now()
	nowCrio := functional.CRIOTime(now)
	logLine := fmt.Sprintf(functional.OpenShiftAuditLogTemplate, nowCrio, nowCrio)
	return logLine, f.WriteMessagesToOpenshiftAuditLog(logLine, 3)
}

func writeOvnAuditLog(f *functional.CollectorFunctionalFramework) (string, error) {
	now := time.Now()
	logLine := functional.NewOVNAuditLog(now)
	return logLine, f.WriteMessagesToOVNAuditLog(logLine, 3)
}

func writeAuditHostLog(f *functional.CollectorFunctionalFramework) (string, error) {
	now := time.Now()
	logLine := functional.NewAuditHostLog(now)
	return logLine, f.WriteMessagesToAuditLog(logLine, 3)
}

func validateHostLog(logLine string, logRecord otlp.LogRecord) {
	parts := strings.Split(logLine, "):")
	parts = strings.Split(parts[0], " ")
	Expect(len(parts)).To(BeNumerically(">", 1), "Not valid test data")
	auditdType := strings.Split(parts[0], "=")[1]
	sequence := strings.Split(parts[1], ":")[1]
	expLogAttributes := map[string]types.GomegaMatcher{
		otlp.AuditdType:     BeEquivalentTo(auditdType),
		otlp.AuditdSequence: BeEquivalentTo(sequence),
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
		otlp.K8sOVNSequence:  BeEquivalentTo(sequence),
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
	expLogAttributes := map[string]types.GomegaMatcher{
		otlp.K8sEventLevel:               BeEquivalentTo(event.Level),
		otlp.K8sEventStage:               BeEquivalentTo(event.Stage),
		otlp.K8sEventUserAgent:           BeEquivalentTo(event.UserAgent),
		otlp.K8sEventRequestURI:          BeEquivalentTo(event.RequestURI),
		otlp.K8sEventResponseCode:        Equal(strconv.Itoa(int(event.ResponseStatus.Code))),
		otlp.K8sEventObjectRefResource:   BeEquivalentTo(event.ObjectRef.Resource),
		otlp.K8sEventObjectRefName:       BeEquivalentTo(event.ObjectRef.Name),
		otlp.K8sEventObjectRefNamespace:  BeEquivalentTo(event.ObjectRef.Namespace),
		otlp.K8sEventObjectRefAPIGroup:   BeEquivalentTo(event.ObjectRef.APIGroup),
		otlp.K8sEventObjectRefAPIVersion: BeEquivalentTo(event.ObjectRef.APIVersion),
		otlp.K8sUserUsername:             BeEquivalentTo(event.User.Username),
	}
	Expect(otlp.CollectNames(logRecord.Attributes)).To(ContainElements(maps.Keys(expLogAttributes)))
	expAttrs := maps.Keys(expLogAttributes)
	sort.Strings(expAttrs)
	for _, key := range expAttrs {
		Expect(logRecord.Attribute(key).String()).To(expLogAttributes[key], key)
	}
	Expect(logRecord.Attribute(otlp.K8sUserGroups).Array.List()).To(ContainElements(event.User.Groups))
	logAttributeNames := otlp.CollectNames(logRecord.Attributes)
	Expect(logAttributeNames).To(ContainElement(HavePrefix(otlp.K8sEventAnnotationPrefix)), "Exp. some attributes with this prefix")
}

func validateOpenshiftAPILogs(logLine string, logRecord otlp.LogRecord) {
	event := &helpertypes.OpenshiftAuditLog{}
	test.MustUnmarshal(logLine, event)
	expLogAttributes := map[string]types.GomegaMatcher{
		otlp.K8sEventLevel:                           BeEquivalentTo(event.Level),
		otlp.K8sEventStage:                           BeEquivalentTo(event.Stage),
		otlp.K8sEventUserAgent:                       BeEquivalentTo(event.UserAgent),
		otlp.K8sEventRequestURI:                      BeEquivalentTo(event.RequestURI),
		otlp.K8sEventResponseCode:                    Equal(strconv.Itoa(event.ResponseStatus.Code)),
		otlp.K8sUserUsername:                         BeEquivalentTo(event.User.Username),
		otlp.K8sEventAnnotationAuthorizationDecision: BeEquivalentTo(event.Annotations.AuthorizationK8SIoDecision),
		otlp.K8sEventAnnotationAuthorizationReason:   BeEquivalentTo(event.Annotations.AuthorizationK8SIoReason),
	}
	Expect(otlp.CollectNames(logRecord.Attributes)).To(ContainElements(maps.Keys(expLogAttributes)))
	expAttrs := maps.Keys(expLogAttributes)
	sort.Strings(expAttrs)
	for _, key := range expAttrs {
		Expect(logRecord.Attribute(key).String()).To(expLogAttributes[key], key)
	}
	Expect(logRecord.Attribute(otlp.K8sUserGroups).Array.List()).To(ContainElements(event.User.Groups))
	logAttributeNames := otlp.CollectNames(logRecord.Attributes)
	Expect(logAttributeNames).To(ContainElement(HavePrefix(otlp.K8sEventAnnotationPrefix)), "Exp. some attributes with this prefix")
}

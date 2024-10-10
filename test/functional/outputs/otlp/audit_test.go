package otlp

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types/otlp"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"golang.org/x/exp/maps"
	"k8s.io/apiserver/pkg/apis/audit"
	"sort"
	"strconv"
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

	Context("When forwarding logs to an otel-collector", func() {
		var (
			openShiftLabels = map[string]string{"foo": "bar"}
		)
		It("should send audit logs with correct grouping and resource attributes", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				WithLabelsFilter(openShiftLabels).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			// Log message data
			now := time.Now()
			k8sAuditLogLine := functional.NewKubeAuditLog(now)
			Expect(framework.WriteMessagesTok8sAuditLog(k8sAuditLogLine, 3)).To(BeNil())

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
				otlp.OpenshiftLogSource:  BeEquivalentTo(obs.AuditSourceKube),
				otlp.OpenshiftLogType:    BeEquivalentTo(obs.InputTypeAudit),
				otlp.NodeName:            MatchRegexp(".*"),
				// deprecated
				otlp.OpenshiftClusterID: MatchRegexp(".*"),
				otlp.KubernetesHost:     MatchRegexp(".*"),
				otlp.LogSource:          BeEquivalentTo(obs.AuditSourceKube),
				otlp.LogType:            BeEquivalentTo(obs.InputTypeAudit),
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

			event := &audit.Event{}
			test.MustUnmarshal(k8sAuditLogLine, event)
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

		})
	})
})

package normalization

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/syslog"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

var _ = Describe("[Functional][Normalization] Messages from EventRouter", func() {

	var (
		framework                          *functional.CollectorFunctionalFramework
		baseTime                           = time.Now().UTC()
		templateForAnyKubernetesWithEvents = types.KubernetesWithEvent{
			Kubernetes: functional.TemplateForAnyKubernetes,
		}
		NewEventDataBuilder = func(verb, message string, podRef *corev1.ObjectReference) types.EventData {
			newEvent := types.NewEvent(podRef, corev1.EventTypeNormal, "reason", message)
			if verb == "UPDATED" {
				oldEvent := types.NewEvent(podRef, corev1.EventTypeWarning, "old_reason", "old_"+message)
				return types.EventData{Verb: "UPDATED", Event: newEvent, OldEvent: oldEvent}
			} else {
				return types.EventData{Verb: "ADDED", Event: newEvent}
			}
		}

		ExpectedLogTemplateBuilder = func(event, oldEvent *corev1.Event) types.EventRouterLog {
			tsTruncated := baseTime.Truncate(time.Second)
			timestamp := tsTruncated
			tmpl := types.EventRouterLog{
				Kubernetes: templateForAnyKubernetesWithEvents,
				ViaQCommon: types.ViaQCommon{
					Message:          event.Message,
					Level:            types.AnyString,
					Hostname:         types.AnyString,
					PipelineMetadata: types.PipelineMetadata{},
					Timestamp:        timestamp,
					TimestampLegacy:  tsTruncated,
					LogSource:        string(obs.InfrastructureSourceContainer),
					LogType:          string(obs.InputTypeApplication),
					Openshift: types.OpenshiftMeta{
						ClusterID: types.AnyString,
						Sequence:  types.NewOptionalInt(""),
					},
				},
			}
			//optional for test given we are mocking and these values may not map to actual meta
			tmpl.Kubernetes.ContainerImage = types.OptionalString
			tmpl.Kubernetes.ContainerImageID = types.OptionalString
			tmpl.Kubernetes.PodID = types.OptionalString
			tmpl.Kubernetes.Event = types.ViaqEventRouterEvent{
				Event: *event,
				Verb:  types.AnyString,
			}
			tmpl.Kubernetes.Event.Message = ""
			if oldEvent != nil {
				tmpl.OldEvent = oldEvent
			}

			return tmpl
		}

		parseLogs = func(raw []string, outputType obs.OutputType) ([]types.EventRouterLog, error) {
			var logs []types.EventRouterLog
			switch outputType {
			case obs.OutputTypeHTTP:
				err := types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
				return logs, err
			case obs.OutputTypeSyslog:
				jsStr := make([]string, len(raw))
				for i, s := range raw {
					s, _ := syslog.ParseRFC5424SyslogLogs(s)
					jsStr[i] = s.MessagePayload
				}
				err := types.StrictlyParseLogs(utils.ToJsonLogs(jsStr), &logs)
				return logs, err
			}
			return nil, nil
		}
	)

	DescribeTable("should be normalized to the ViaQ data model when sinking to different outputs", func(outputType obs.OutputType, verb, message, expectedMessage string) {
		crioTimestamp := functional.CRIOTime(baseTime)
		framework = functional.NewCollectorFunctionalFramework()
		builder := testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication)
		if outputType == obs.OutputTypeHTTP {
			builder.ToHttpOutput()
		}
		if outputType == obs.OutputTypeSyslog {
			builder.ToSyslogOutput(obs.SyslogRFC5424)
		}

		framework.VisitConfig = func(conf string) string {
			return strings.Replace(conf, `"eventrouter-"`, `"functional"`, 1)
		}
		Expect(framework.Deploy()).To(BeNil())
		podRef, err := reference.GetReference(scheme.Scheme, types.NewMockPod())
		Expect(err).To(BeNil())
		newEventData := NewEventDataBuilder(verb, message, podRef)
		tsSec := baseTime.Truncate(time.Second)
		newEventData.Event.FirstTimestamp = metav1.NewTime(tsSec)
		newEventData.Event.LastTimestamp = metav1.NewTime(tsSec)
		if verb == "UPDATED" {
			tsOld := tsSec.Add(-10 * time.Minute)
			newEventData.Event.FirstTimestamp = metav1.NewTime(tsOld)
			newEventData.OldEvent.FirstTimestamp = metav1.NewTime(tsOld)
			newEventData.OldEvent.LastTimestamp = metav1.NewTime(tsSec)
		}
		jsonBytes, _ := json.Marshal(newEventData)
		jsonStr := string(jsonBytes)
		msg := functional.NewCRIOLogMessage(crioTimestamp, jsonStr, false)
		err = framework.WriteMessagesToApplicationLog(msg, 1)
		Expect(err).To(BeNil())

		raw, err := framework.ReadRawApplicationLogsFrom(string(outputType))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := parseLogs(raw, outputType)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		expectedEventData := newEventData
		expectedEventData.Event.Message = expectedMessage
		var expectedLogTemplate = ExpectedLogTemplateBuilder(expectedEventData.Event, expectedEventData.OldEvent)
		Expect(logs[0]).To(matchers.FitLogFormatTemplate(expectedLogTemplate))
	},
		Entry("with HTTP output for ADDED events", obs.OutputTypeHTTP, "ADDED", "simple syslog message", "simple syslog message"),
		Entry("with HTTP output for UPDATED events", obs.OutputTypeHTTP, "UPDATED", "simple syslog message", "simple syslog message"),
		Entry("with Syslog output for ADDED events", obs.OutputTypeSyslog, "ADDED", "simple syslog message", "simple syslog message"),
		Entry("with Syslog output for UPDATED events", obs.OutputTypeSyslog, "UPDATED", "simple syslog message", "simple syslog message"),
		Entry("with Syslog output for ADDED events and new line symbol", obs.OutputTypeSyslog, "ADDED", "syslog message\n with new line", "syslog message\\n with new line"),
		Entry("with Syslog output for UPDATED events and new line symbol", obs.OutputTypeSyslog, "UPDATED", "syslog message\n with new line", "syslog message\\n with new line"),
	)

	DescribeTable("should use the correct timestamp fallback for @timestamp", func(setTimestamps func(event *corev1.Event, ts time.Time)) {
		crioTimestamp := functional.CRIOTime(baseTime)
		framework = functional.NewCollectorFunctionalFramework()
		builder := testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication)
		builder.ToHttpOutput()

		framework.VisitConfig = func(conf string) string {
			return strings.Replace(conf, `"eventrouter-"`, `"functional"`, 1)
		}
		Expect(framework.Deploy()).To(BeNil())
		podRef, err := reference.GetReference(scheme.Scheme, types.NewMockPod())
		Expect(err).To(BeNil())
		newEventData := NewEventDataBuilder("ADDED", "test message", podRef)
		setTimestamps(newEventData.Event, baseTime.Truncate(time.Second))
		jsonBytes, _ := json.Marshal(newEventData)
		msg := functional.NewCRIOLogMessage(crioTimestamp, string(jsonBytes), false)
		err = framework.WriteMessagesToApplicationLog(msg, 1)
		Expect(err).To(BeNil())

		raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := parseLogs(raw, obs.OutputTypeHTTP)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs[0].TimestampLegacy.IsZero()).To(BeFalse(), "Expected @timestamp to be set via timestamp fallback chain")
	},
		Entry("lastTimestamp is set", func(event *corev1.Event, ts time.Time) {
			event.LastTimestamp = metav1.NewTime(ts)
			event.FirstTimestamp = metav1.NewTime(ts.Add(-5 * time.Minute))
		}),
		Entry("lastTimestamp is zero, falls back to firstTimestamp", func(event *corev1.Event, ts time.Time) {
			event.FirstTimestamp = metav1.NewTime(ts)
		}),
		Entry("lastTimestamp and firstTimestamp are zero, falls back to eventTime", func(event *corev1.Event, ts time.Time) {
			event.EventTime = metav1.NewMicroTime(ts)
		}),
		Entry("all timestamps zero, falls back to creationTimestamp", func(event *corev1.Event, ts time.Time) {
			event.CreationTimestamp = metav1.NewTime(ts)
		}),
	)

	AfterEach(func() {
		if framework != nil {
			framework.Cleanup()
		}
	})
})

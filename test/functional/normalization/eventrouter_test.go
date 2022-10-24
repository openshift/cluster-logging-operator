//go:build fluentd
// +build fluentd

package normalization

import (
	"encoding/json"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

var _ = Describe("[Functional][Normalization] Fluentd normalization for EventRouter messages", func() {

	const timestamp string = "1985-10-21T09:00:00.00000+00:00"
	var (
		framework                *functional.CollectorFunctionalFramework
		pod                      *corev1.Pod
		nanoTime, _              = time.Parse(time.RFC3339Nano, timestamp)
		templateForAnyKubernetes = types.Kubernetes{
			ContainerName:    "*",
			NamespaceName:    "*",
			PodName:          "*",
			ContainerImage:   "*",
			ContainerImageID: "*",
			PodID:            "*",
			PodIP:            "**optional**",
			Host:             "*",
			MasterURL:        "*",
			NamespaceID:      "*",
			FlatLabels:       []string{"*"},
			NamespaceLabels:  map[string]string{"*": "*"},
		}
		templateForAnyCollector = types.PipelineMetadata{
			Collector: types.Collector{
				Ipaddr4:    "*",
				Inputname:  "*",
				Name:       "*",
				Version:    "*",
				ReceivedAt: time.Time{},
			},
		}
		NewEventDataBuilder = func(verb string, podRef *corev1.ObjectReference) types.EventData {
			newEvent := types.NewMockEvent(podRef, corev1.EventTypeNormal, "reason", "message")
			if verb == "UPDATED" {
				oldEvent := types.NewMockEvent(podRef, corev1.EventTypeWarning, "old_reason", "old_message")
				return types.EventData{Verb: "UPDATED", Event: newEvent, OldEvent: oldEvent}
			} else {
				return types.EventData{Verb: "ADDED", Event: newEvent}
			}
		}

		ExpectedLogTemplateBuilder = func(message string, timestamp time.Time) types.EventRouterLog {
			return types.EventRouterLog{
				Docker: types.Docker{
					ContainerID: "*",
				},
				Kubernetes:       templateForAnyKubernetes,
				Message:          message,
				Level:            "*",
				Hostname:         "*",
				PipelineMetadata: templateForAnyCollector,
				Timestamp:        timestamp,
				LogType:          "application",
				ViaqMsgID:        "*",
				OpenshiftLabels: types.OpenshiftMeta{
					ClusterID: "*",
					Sequence:  types.NewOptionalInt(""),
				},
			}
		}
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
		pod = types.NewMockPod()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("when normalizing events", func(verb string) {
		podRef, err := reference.GetReference(scheme.Scheme, pod)
		Expect(err).To(BeNil())
		newEventData := NewEventDataBuilder(verb, podRef)
		jsonBytes, _ := json.Marshal(newEventData)
		jsonStr := string(jsonBytes)
		msg := functional.NewCRIOLogMessage(timestamp, jsonStr, false)
		err = framework.WriteMessagesToApplicationLog(msg, 1)
		Expect(err).To(BeNil())

		raw, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.EventRouterLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		var expectedLogTemplate = ExpectedLogTemplateBuilder(jsonStr, nanoTime)
		outputTestLog := logs[0]
		Expect(outputTestLog).To(matchers.FitLogFormatTemplate(expectedLogTemplate))
	},
		Entry("It should normalize 'ADDED' events", "ADDED"),
		Entry("It should normalize 'UPDATED' events", "UPDATED"),
	)

})

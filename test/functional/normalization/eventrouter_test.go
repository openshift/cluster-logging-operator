package normalization

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	"strings"
	"time"
)

const timestamp string = "1985-10-21T09:00:00.00000+00:00"

var (
	templateForAnyKubernetes = types.Kubernetes{
		ContainerName:     "*",
		PodName:           "*",
		NamespaceName:     "*",
		NamespaceID:       "*",
		OrphanedNamespace: "*",
	}
	templateForAnyCollector = types.PipelineMetadata{Collector: types.Collector{
		Ipaddr4:    "*",
		Inputname:  "*",
		Name:       "*",
		Version:    "*",
		ReceivedAt: time.Time{},
	},
	}
	nanoTime, _ = time.Parse(time.RFC3339Nano, timestamp)
)
var _ = Describe("[Normalization] Fluentd normalization for EventRouter messages", func() {

	var (
		framework *functional.FluentdFunctionalFramework
		pod       *corev1.Pod
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
		pod = types.NewMockPod()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("Should parse EventRouter `ADDED` message and check values", func() {
		podRef, err := reference.GetReference(scheme.Scheme, pod)
		Expect(err).To(BeNil())
		newEvent := types.NewMockEvent(podRef,
			corev1.EventTypeNormal,
			string(utils.GetRandomWord(8)),
			string(utils.GetRandomWord(32)))
		newEventData := types.EventData{Verb: "ADDED", Event: newEvent}
		jsonBytes, _ := json.Marshal(newEventData)
		jsonStr := string(jsonBytes)
		msg := strings.ReplaceAll(fmt.Sprintf("%s stdout F %s", timestamp, jsonStr), "\"", "\\\"")
		err = framework.WriteMessagesToApplicationLog(msg, 1)
		Expect(err).To(BeNil())

		raw, err := framework.ReadApplicationLogsFrom("fluentforward")
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.EventRouterLog
		err = types.StrictlyParseLogs(raw, &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		var outputLogTemplate = OutPutTemplate(jsonStr)
		outputTestLog := logs[0]

		Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
	})

	It("Should parse EventRouter `UPDATED` message and check values", func() {
		podRef, err := reference.GetReference(scheme.Scheme, pod)
		Expect(err).To(BeNil())
		newEvent := types.NewMockEvent(podRef,
			corev1.EventTypeNormal,
			string(utils.GetRandomWord(8)),
			string(utils.GetRandomWord(32)))
		oldEvent := types.NewMockEvent(podRef,
			corev1.EventTypeWarning,
			string(utils.GetRandomWord(8)),
			string(utils.GetRandomWord(32)))
		newEventData := types.EventData{Verb: "UPDATED", Event: newEvent, OldEvent: oldEvent}
		jsonBytes, _ := json.Marshal(newEventData)
		jsonStr := string(jsonBytes)
		msg := strings.ReplaceAll(fmt.Sprintf("%s stdout F %s", timestamp, jsonStr), "\"", "\\\"")
		err = framework.WriteMessagesToApplicationLog(msg, 1)
		Expect(err).To(BeNil())

		raw, err := framework.ReadApplicationLogsFrom("fluentforward")
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.EventRouterLog
		err = types.StrictlyParseLogs(raw, &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		var outputLogTemplate = OutPutTemplate(jsonStr)
		outputTestLog := logs[0]

		Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
	})
})

func OutPutTemplate(message string) types.EventRouterLog {
	return types.EventRouterLog{
		Docker: types.Docker{
			ContainerID: "*",
		},
		Kubernetes:       templateForAnyKubernetes,
		Message:          message,
		Level:            "unknown",
		PipelineMetadata: templateForAnyCollector,
		Timestamp:        nanoTime,
		ViaqIndexName:    "app-write",
		ViaqMsgID:        "*",
		OpenshiftLabels:  types.OpenshiftMeta{},
	}
}

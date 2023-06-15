package normalization

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	"strings"
)

var _ = Describe("[Functional][Normalization] Fluentd normalization for EventRouter messages", func() {

	const timestamp string = "2013-03-28T14:36:03.243000+00:00"

	var (
		fluentdLogPath = map[string]string{
			"application": "/var/log/pods",
		}
		framework *functional.CollectorFunctionalFramework
		pod       *corev1.Pod
		//nanoTime, _              = time.Parse(time.RF, timestamp)
		templateForInfraKubernetes = types.Kubernetes{
			ContainerID:       "**optional**",
			ContainerName:     "*",
			PodName:           "*",
			NamespaceName:     "*",
			NamespaceID:       "**optional**",
			OrphanedNamespace: "**optional**",
			ContainerImage:    "**optional**",
			ContainerImageID:  "**optional**",
			PodID:             "**optional**",
			PodIP:             "**optional**",
			Host:              "**optional**",
			MasterURL:         "**optional**",
			FlatLabels:        []string{"*"},
			NamespaceLabels:   map[string]string{"*": "*"},
			Annotations:       map[string]string{"*": "*"},
		}
		eventDataBuilder = func(verb string, podRef *corev1.ObjectReference) string {
			if verb == "ADDED" {
				return fmt.Sprintf("{\"verb\":\"%s\",\"event\":{\"metadata\":{\"name\":\"%s\",\"namespace\":\"%s\","+
					"\"creationTimestamp\":\"2013-03-28T14:36:03.243000+00:00\"},\"involvedObject\":{\"kind\":\"%s\","+
					"\"namespace\":\"%s\",\"name\":\"%s\",\"uid\":\"%s\"},\"reason\":\"reason\","+
					"\"message\":\"message\",\"source\":{},\"firstTimestamp\":\"2013-03-28T14:36:03.243000+00:00\","+
					"\"lastTimestamp\":\"2013-03-28T14:36:03.243000+00:00\",\"count\":2052921995,\"type\":\"Normal\","+
					"\"eventTime\":\"2013-03-28T14:36:03.243000+00:00\",\"reportingComponent\":\"\",\"reportingInstance\":\"\"}}",
					verb, podRef.Name, podRef.Namespace, podRef.Kind, podRef.Namespace, podRef.Name, podRef.UID)
			} else {
				return "{\"verb\":\"UPDATED\",\"event\":{\"metadata\":{\"namespace\":\"6G7smh2R\",\"creationTimestamp\":\"2013-03-28T14:36:03.243000+00:00\"}," +
					"\"involvedObject\":{\"kind\":\"Pod\",\"namespace\":\"6G7smh2R\",\"name\":\"4Jy4pi7R\",\"uid\":\"xAa706rMdS7xhFDT\"}," +
					"\"reason\":\"reason\",\"message\":\"message\",\"source\":{},\"firstTimestamp\":\"2013-03-28T14:36:03.243000+00:00\"," +
					"\"lastTimestamp\":\"2013-03-28T14:36:03.243000+00:00\",\"count\":258172305,\"type\":\"Normal\",\"eventTime\":\"2013-03-28T14:36:03.243000+00:00\"," +
					"\"reportingComponent\":\"\",\"reportingInstance\":\"\"},\"old_event\":{\"metadata\":{\"namespace\":\"6G7smh2R\"," +
					"\"creationTimestamp\":\"2013-03-28T14:36:03.243000+00:00\"},\"involvedObject\":{\"kind\":\"Pod\",\"namespace\":\"6G7smh2R\",\"name\":\"4Jy4pi7R\"," +
					"\"uid\":\"xAa706rMdS7xhFDT\"},\"reason\":\"old_reason\",\"message\":\"old_message\",\"source\":{}," +
					"\"firstTimestamp\":\"2013-03-28T14:36:03.243000+00:00\",\"lastTimestamp\":\"2013-03-28T14:36:03.243000+00:00\",\"count\":126878443," +
					"\"type\":\"Warning\",\"eventTime\":\"2013-03-28T14:36:03.243000+00:00\",\"reportingComponent\":\"\",\"reportingInstance\":\"\"}}"
			}
		}

		expectedLogTemplateBuilder = func(verb, message, timestamp string) types.EventRouterLog {
			if verb == "ADDED" {
				return types.EventRouterLog{
					ViaqMsgID:  "**optional**",
					Kubernetes: templateForInfraKubernetes,
					Docker: types.Docker{
						ContainerID: "**optional**",
					},
					PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
					Message:          message,
					Level:            "*",
					Hostname:         "*",
					Timestamp:        "*",
					LogType:          "infrastructure",
					OpenshiftLabels: types.OpenshiftMeta{
						ClusterID: "*",
						Sequence:  types.NewOptionalInt(""),
					},
					Event: corev1.Event{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "**optional**",
							Namespace: "*",
						},
						EventTime:           metav1.MicroTime{},
						Series:              nil,
						ReportingController: "",
						ReportingInstance:   "",
						Reason:              "*",
						Type:                "*",
						InvolvedObject: corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "*",
							Name:      "*",
							UID:       "*",
						},
						Message: "*",
					},
					Verb: "*",
				}
			} else {
				return types.EventRouterLog{
					ViaqMsgID:  "**optional**",
					Kubernetes: templateForInfraKubernetes,
					Docker: types.Docker{
						ContainerID: "**optional**",
					},
					PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
					Message:          message,
					Level:            "*",
					Hostname:         "*",
					Timestamp:        "*",
					LogType:          "infrastructure",
					OpenshiftLabels: types.OpenshiftMeta{
						ClusterID: "*",
						Sequence:  types.NewOptionalInt(""),
					},
					Event: corev1.Event{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "**optional**",
							Namespace: "*",
						},
						EventTime:           metav1.MicroTime{},
						Series:              nil,
						ReportingController: "",
						ReportingInstance:   "",
						Reason:              "*",
						Type:                "*",
						InvolvedObject: corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "*",
							Name:      "*",
							UID:       "*",
						},
						Message: "*",
					},
					OldEvent: corev1.Event{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "**optional**",
							Namespace: "*",
						},
						EventTime:           metav1.MicroTime{},
						Series:              nil,
						ReportingController: "",
						ReportingInstance:   "",
						Reason:              "*",
						Type:                "*",
						InvolvedObject: corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "*",
							Name:      "*",
							UID:       "*",
						},
						Message: "*",
					},
					Verb: "*",
				}
			}
		}
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType, client.UseInfraNamespaceTestOption)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameInfrastructure).
			ToElasticSearchOutput().
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput()
		framework.VisitConfig = func(conf string) string {
			switch testfw.LogCollectionType {
			case logging.LogCollectionTypeFluentd:
				conf = strings.Replace(conf, "@type kubernetes_metadata", "@type kubernetes_metadata\ntest_api_adapter  KubernetesMetadata::TestApiAdapter\n", 1)
			//conf = strings.Replace(conf, "alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'", "alt_tags 'kubernetes.var.log.pods.**_even_trou_ter-*.** kubernetes.journal.container._default_.kubernetes-event'", 1)
			case logging.LogCollectionTypeVector:
				conf = strings.Replace(conf, fmt.Sprintf("exclude_paths_glob_patterns = [\"/var/log/pods/%s_collector-*/*/*.log\",", framework.Namespace), "exclude_paths_glob_patterns = [", 1)
				//conf = strings.Replace(conf, "alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'", "alt_tags 'kubernetes.var.log.pods.**_even_trou_ter-*.** kubernetes.journal.container._default_.kubernetes-event'", 1)
			}
			//exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log"
			return conf
		}
		Expect(framework.Deploy()).To(BeNil())
		pod = types.NewMockPod()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("when normalizing events", func(verb string) {
		podRef, err := reference.GetReference(scheme.Scheme, pod)
		Expect(err).To(BeNil())
		jsonStr := eventDataBuilder(verb, podRef)
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> jsonStr")
		fmt.Println(jsonStr)
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		msg := functional.NewCRIOLogMessage(timestamp, jsonStr, false)

		if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
			ns := "openshift-fake-infra"
			if strings.HasPrefix(framework.Namespace, "openshift-test") {
				ns = framework.Namespace
			}
			filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath["application"], ns, "eventrouter-ghg", framework.Pod.UID, constants.CollectorName)
			err = framework.WriteMessagesToLog(msg, 1, filename)
		} else {
			err = framework.WriteMessagesToApplicationLog(msg, 1)
			Expect(err).To(BeNil())
		}
		var raw []string
		if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
			raw, err = framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, "infra-write")
		} else {
			raw, err = framework.ReadInfrastructureLogsFrom(logging.OutputTypeElasticsearch)
		}

		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> raw")
		fmt.Println(raw)
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")

		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.EventRouterLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		var expectedLogTemplate = expectedLogTemplateBuilder(verb, jsonStr, timestamp)
		outputTestLog := logs[0]
		Expect(outputTestLog).To(matchers.FitLogFormatTemplate(expectedLogTemplate))
	},
		Entry("It should normalize 'ADDED' events", "ADDED"),
		Entry("It should normalize 'UPDATED' events", "UPDATED"),
	)

})

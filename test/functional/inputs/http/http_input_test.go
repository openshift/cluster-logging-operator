package http

import (
	"encoding/json"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

const (
	httpInputName  = `http-source`
	servicePortNum = 8080

	// eventsString is a sample from the API-server webhook.
	// It does not include the 'apiVersion' and 'kind' fields in each event.
	// Use this exact serialization for testing to ensure compatibility with the real API server.
	// For more see: https://issues.redhat.com/browse/LOG-4681
	eventsString = `{
  "apiVersion": "audit.k8s.io/v1",
  "items": [
    {
      "annotations": {
        "authorization.k8s.io/decision": "allow",
        "authorization.k8s.io/reason": ""
      },
      "auditID": "0c8bc986-b74c-4392-8756-7bdbaf3cc63e",
      "level": "Metadata",
      "objectRef": {
        "apiGroup": "user.openshift.io",
        "apiVersion": "v1",
        "resource": "groups"
      },
      "requestReceivedTimestamp": "2023-10-18T09:26:15.332127Z",
      "requestURI": "/apis/user.openshift.io/v1/groups?allowWatchBookmarks=true&resourceVersion=205534&timeout=5m15s&timeoutSeconds=315&watch=true",
      "responseStatus": {
        "code": 200,
        "metadata": {}
      },
      "sourceIPs": [
        "::1"
      ],
      "stage": "ResponseComplete",
      "stageTimestamp": "2023-10-18T09:31:30.332784Z",
      "user": {
        "groups": [
          "system:masters"
        ],
        "uid": "e71d8fb7-531f-4b8d-bdba-65abb9c0849b",
        "username": "system:apiserver"
      },
      "userAgent": "oauth-apiserver/v0.0.0 (linux/amd64) kubernetes/$Format",
      "verb": "watch"
    },
    {
      "annotations": {
        "authorization.k8s.io/decision": "allow",
        "authorization.k8s.io/reason": ""
      },
      "auditID": "3803e946-cbd2-4fe7-b8c0-57dba6a849d9",
      "level": "Metadata",
      "objectRef": {
        "apiGroup": "user.openshift.io",
        "apiVersion": "v1",
        "resource": "groups"
      },
      "requestReceivedTimestamp": "2023-10-18T09:31:30.333236Z",
      "requestURI": "/apis/user.openshift.io/v1/groups?allowWatchBookmarks=true&resourceVersion=207810&timeout=8m15s&timeoutSeconds=495&watch=true",
      "responseStatus": {
        "code": 200,
        "metadata": {}
      },
      "sourceIPs": [
        "::1"
      ],
      "stage": "ResponseStarted",
      "stageTimestamp": "2023-10-18T09:31:30.333626Z",
      "user": {
        "groups": [
          "system:masters"
        ],
        "uid": "e71d8fb7-531f-4b8d-bdba-65abb9c0849b",
        "username": "system:apiserver"
      },
      "userAgent": "oauth-apiserver/v0.0.0 (linux/amd64) kubernetes/$Format",
      "verb": "watch"
    }
  ],
  "kind": "EventList",
  "metadata": {},
  "path": "/",
  "source_type": "http_server",
  "timestamp": "2023-10-18T09:31:53.359093622Z"
}`
)

var _ = Describe("[Functional][Inputs][Http] Functional tests", func() {

	if testfw.LogCollectionType != logging.LogCollectionTypeVector {
		defer GinkgoRecover()
		Skip("skip for non-vector")
	}

	var (
		framework *functional.CollectorFunctionalFramework
		events    auditv1.EventList
	)
	eventsBytes := []byte(strings.Replace(eventsString, "\n", "", -1))

	BeforeEach(func() {
		Expect(json.Unmarshal(eventsBytes, &events)).To(Succeed())
		Expect(testfw.LogCollectionType).To(Equal(logging.LogCollectionTypeVector))
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		framework.VisitConfig = func(conf string) string {
			return strings.Replace(conf, "enabled = true", "enabled = false", 2) // turn off TLS for testing
		}
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInputWithVisitor(httpInputName,
				func(spec *logging.InputSpec) {
					spec.Receiver = &logging.ReceiverSpec{
						Type: logging.ReceiverTypeHttp,
						ReceiverTypeSpec: &logging.ReceiverTypeSpec{
							HTTP: &logging.HTTPReceiver{
								Port:   servicePortNum,
								Format: logging.FormatKubeAPIAudit,
							},
						},
					}
				}).ToHttpOutput()
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When sending an audit log record to an HTTP input", func() {
		It("should be able to round trip it unharmed", func() {
			Expect(framework.DeployWithVisitor(
				func(b *runtime.PodBuilder) error {
					return framework.AddVectorHttpOutput(b, framework.Forwarder.Spec.Outputs[0])
				}),
			).To(BeNil())
			err := framework.WriteAsJsonToHttpInput(httpInputName, events.Items[0])
			Expect(err).To(BeNil(), "Expected no errors writing to HTTP input")
			raw, err := framework.ReadFileFromWithRetryInterval("http", functional.ApplicationLogFile, time.Second)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			var logs []auditv1.Event
			err = types.ParseLogsFrom(utils.ToJsonLogs([]string{raw}), &logs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := logs[0]
			Expect(outputTestLog).To(FitLogFormatTemplate(events.Items[0]), "raw string: %q", raw)
		})
	})

	Context("When sending an array of audit log records to an HTTP input", func() {
		It("should be able to round trip them unharmed", func() {
			Expect(framework.DeployWithVisitor(
				func(b *runtime.PodBuilder) error {
					return framework.AddVectorHttpOutput(b, framework.Forwarder.Spec.Outputs[0])
				}),
			).To(BeNil())

			err := framework.WriteToHttpInputWithPortForwarder(httpInputName, eventsBytes)
			Expect(err).To(BeNil(), "Expected no errors writing to HTTP input")
			raw, err := framework.ReadFileFromWithRetryInterval("http", functional.ApplicationLogFile, time.Second)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			lines := strings.Split(strings.TrimSpace(raw), "\n")
			Expect(len(lines)).To(Equal(len(events.Items)), "--- raw lines:\n%v\n...", raw)
			for i, line := range lines {
				var logs []auditv1.Event
				err = types.ParseLogsFrom(utils.ToJsonLogs([]string{line}), &logs, false)
				Expect(err).To(BeNil(), "Expected no errors parsing the logs")
				Expect(logs[0]).To(FitLogFormatTemplate(events.Items[i]))
			}
		})

		It("should apply an audit filter", func() {
			Expect(events.Items[0].Stage).To(Not(Equal(events.Items[1].Stage)))
			filterName := "auditFilter"
			framework.Forwarder.Spec.Filters = []logging.FilterSpec{{
				Name: filterName,
				Type: logging.FilterKubeAPIAudit,
				FilterTypeSpec: logging.FilterTypeSpec{KubeAPIAudit: &logging.KubeAPIAudit{
					OmitStages: []auditv1.Stage{events.Items[0].Stage},
					Rules:      []auditv1.PolicyRule{{Level: auditv1.LevelMetadata}}, // Skip default rules.
				}},
			}}
			framework.Forwarder.Spec.Pipelines[0].FilterRefs = []string{filterName}

			Expect(framework.DeployWithVisitor(
				func(b *runtime.PodBuilder) error {
					return framework.AddVectorHttpOutput(b, framework.Forwarder.Spec.Outputs[0])
				}),
			).To(BeNil())

			err := framework.WriteToHttpInputWithPortForwarder(httpInputName, eventsBytes)
			Expect(err).To(BeNil(), "Expected no errors writing to HTTP input")
			raw, err := framework.ReadFileFromWithRetryInterval("http", functional.ApplicationLogFile, time.Second)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			lines := strings.Split(strings.TrimSpace(raw), "\n")
			Expect(lines).To(HaveLen(1))
			var logs []auditv1.Event
			err = types.ParseLogsFrom(utils.ToJsonLogs(lines), &logs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(logs[0]).To(FitLogFormatTemplate(events.Items[1]))
		})
	})
})

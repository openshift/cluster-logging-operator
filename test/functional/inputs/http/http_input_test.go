// go:build !fluentd
package http

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	authentication "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

const (
	httpInputName  = `http-source`
	servicePortNum = 8080
)

var _ = Describe("[Functional][Inputs][Http] Functional tests", func() {

	if testfw.LogCollectionType != logging.LogCollectionTypeVector {
		defer GinkgoRecover()
		Skip("skip for non-vector")
	}

	tsF := func(tsStr string) time.Time { ts, _ := time.Parse(time.RFC3339Nano, tsStr); return ts }

	var (
		framework *functional.CollectorFunctionalFramework

		auditRecord = auditv1.Event{
			TypeMeta: metav1.TypeMeta{
				Kind:       `Event`,
				APIVersion: `audit.k8s.io/v1`,
			},
			Level:      `default`,
			AuditID:    `16dcc977-a39e-467b-af44-921416b7800b`,
			Stage:      `RequestReceived`,
			RequestURI: `/apis/rbac.authorization.k8s.io/v1/clusterrolebindings?limit=500\u0026resourceVersion=0`,
			Verb:       `create`,
			User: authentication.UserInfo{
				Username: "foobar",
				Groups:   []string{"system:masters", "system:authenticated"},
			},
			SourceIPs: []string{"::1", "127.0.0.1"},
			UserAgent: "kube-apiserver/v1.27.3 (linux/amd64) kubernetes/25b4e43",
			ObjectRef: &auditv1.ObjectReference{
				Resource:   "clusterrolebindings",
				APIGroup:   "rbac.authorization.k8s.io",
				APIVersion: "v1",
			},
			StageTimestamp:           metav1.NewMicroTime(tsF(`2022-08-17T20:27:20.570375Z`)),
			RequestReceivedTimestamp: metav1.NewMicroTime(tsF(`2023-08-24T21:37:38.842649Z`)),
			Annotations:              map[string]string{"foo": "bar"},
		}
	)

	BeforeEach(func() {
		Expect(testfw.LogCollectionType).To(Equal(logging.LogCollectionTypeVector))
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		framework.VisitConfig = func(conf string) string {
			return strings.Replace(conf, "enabled = true", "enabled = false", 2) // turn off TLS for testing
		}
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInputWithVisitor(httpInputName,
				func(spec *logging.InputSpec) {
					spec.Receiver = &logging.ReceiverSpec{
						HTTP: &logging.HTTPReceiver{
							ReceiverPort: logging.ReceiverPort{
								Port: servicePortNum,
							},
							Format: logging.FormatKubeAPIAudit,
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
			err := framework.WriteAsJsonToHttpInput(httpInputName, auditRecord)
			Expect(err).To(BeNil(), "Expected no errors writing to HTTP input")
			raw, err := framework.ReadFileFromWithRetryInterval("http", functional.ApplicationLogFile, time.Second)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			var logs []auditv1.Event
			err = types.ParseLogsFrom(utils.ToJsonLogs([]string{raw}), &logs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := logs[0]
			Expect(outputTestLog).To(FitLogFormatTemplate(auditRecord))
		})
	})

	Context("When sending an array of audit log records to an HTTP input", func() {
		It("should be able to round trip them unharmed", func() {
			Expect(framework.DeployWithVisitor(
				func(b *runtime.PodBuilder) error {
					return framework.AddVectorHttpOutput(b, framework.Forwarder.Spec.Outputs[0])
				}),
			).To(BeNil())

			var logArray = auditv1.EventList{
				Items: []auditv1.Event{auditRecord, auditRecord},
			}

			err := framework.WriteAsJsonToHttpInput(httpInputName, logArray)
			Expect(err).To(BeNil(), "Expected no errors writing to HTTP input")
			raw, err := framework.ReadFileFromWithRetryInterval("http", functional.ApplicationLogFile, time.Second)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			lines := strings.Split(raw, "\n")
			Expect(len(lines)).To(Equal(len(logArray.Items)))
			for _, line := range lines {
				var logs []auditv1.Event
				err = types.ParseLogsFrom(utils.ToJsonLogs([]string{line}), &logs, false)
				Expect(err).To(BeNil(), "Expected no errors parsing the logs")
				Expect(logs[0]).To(FitLogFormatTemplate(auditRecord))
			}
		})
	})
})

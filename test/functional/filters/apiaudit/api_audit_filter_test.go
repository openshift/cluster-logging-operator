package apiaudit

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var (
	//go:embed testdata/audit.log
	eventsIn []byte
	//go:embed testdata/audit-want.log
	auditWantLog string
)

var _ = Describe("API audit filter", func() {
	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		l = loki.NewReceiver(f.Namespace, "loki-server")
		Expect(l.Create(f.Test.Client)).To(Succeed())

		// Set up the common template forwarder configuration.
		f.Forwarder.Spec.Outputs = append(f.Forwarder.Spec.Outputs,
			logging.OutputSpec{
				Name:           logging.OutputTypeLoki,
				Type:           logging.OutputTypeLoki,
				URL:            l.InternalURL("").String(),
				OutputTypeSpec: logging.OutputTypeSpec{Loki: &logging.Loki{}},
			})
		f.Forwarder.Spec.Filters = []logging.FilterSpec{
			{
				Name: "my-audit",
				Type: loggingv1.FilterKubeAPIAudit,
				FilterTypeSpec: loggingv1.FilterTypeSpec{
					KubeAPIAudit: &loggingv1.KubeAPIAudit{
						Rules: []auditv1.PolicyRule{
							{Level: auditv1.LevelRequestResponse, Users: []string{"*apiserver"}}, // Keep full event for user ending in *apiserver
							{Level: auditv1.LevelNone, Verbs: []string{"get"}},                   // Drop other get requests
							{Level: auditv1.LevelRequest, Verbs: []string{"patch"}},              // Request data for patch requests
							{Level: auditv1.LevelMetadata},                                       // Metadata for everything else.
						},
					},
				},
			},
		}
		f.Forwarder.Spec.Pipelines = append(f.Forwarder.Spec.Pipelines,
			logging.PipelineSpec{
				Name:       "functional-loki-pipeline_0_",
				FilterRefs: []string{"my-audit"},
				OutputRefs: []string{logging.OutputTypeLoki},
				InputRefs:  []string{logging.InputNameAudit},
			})
	})

	AfterEach(func() {
		f.Cleanup()
	})

	It("should filter events as expected", func() {
		decode := func(eventsJson []string) (events []auditv1.Event) {
			array := fmt.Sprintf("[%v]", strings.Join(eventsJson, ","))
			Expect(json.Unmarshal([]byte(array), &events)).To(Succeed())
			return events
		}
		Expect(f.Deploy()).To(Succeed())
		Expect(f.WriteLog(filepath.Join(functional.K8sAuditLogDir, "audit.log"), eventsIn)).To(Succeed())
		want := decode(strings.Split(strings.TrimSpace(string(auditWantLog)), "\n"))

		// Get actual events from Loki
		result, err := l.QueryUntil(`{log_type="audit"}`, "", len(want))
		Expect(err).To(Succeed())
		got := decode(result[0].Lines())
		Expect(got).To(EqualDiff(want))
	})
})

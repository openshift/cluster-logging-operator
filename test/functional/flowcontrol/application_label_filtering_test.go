// go:build !fluentd
package flowcontrol

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[Functional][FlowControl] Policies at Input", func() {
	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)
	appLabels := map[string]string{"name": "app1", "env": "env1"}

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		l = loki.NewReceiver(f.Namespace, "loki-server")
		Expect(l.Create(f.Test.Client)).To(Succeed())

		f.Forwarder.Spec.Inputs = append(f.Forwarder.Spec.Inputs,
			logging.InputSpec{
				Name: "custom-app",
				Application: &logging.Application{
					Selector: &logging.LabelSelector{
						MatchLabels: appLabels,
					},
				},
			},
		)

		// Set up the common template forwarder configuration.
		f.Forwarder.Spec.Outputs = append(f.Forwarder.Spec.Outputs,
			logging.OutputSpec{
				Name: logging.OutputTypeLoki,
				Type: logging.OutputTypeLoki,
				URL:  l.InternalURL("").String(),
				OutputTypeSpec: logging.OutputTypeSpec{
					Loki: &logging.Loki{},
				},
			})

		f.Forwarder.Spec.Pipelines = append(f.Forwarder.Spec.Pipelines,
			logging.PipelineSpec{
				InputRefs:  []string{"custom-app"},
				OutputRefs: []string{logging.OutputTypeLoki},
				Name:       "flow-control",
			},
		)

		f.Labels = map[string]string{
			"name": "app1",
			"env":  "env1",
		}
	})

	AfterEach(func() {
		f.Cleanup()
	})
	Describe("when Source is Application", func() {
		Describe("rate limiting logs by label selector", func() {
			It("applying drop policy at group level", func() {
				if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
					Skip("Skipping test since flow-control is not supported with fluentd")
				}

				f.Forwarder.Spec.Inputs[0].Application.GroupLimit = &logging.LimitSpec{
					Policy:              logging.DropPolicy,
					MaxRecordsPerSecond: 10,
				}

				Expect(f.Deploy()).To(BeNil())
				Expect(WriteApplicationLogs(f, 1000)).To(Succeed())

				r, err := l.QueryUntil(fmt.Sprintf(LokiNsQuery, AllLogs), "", 10)
				Expect(err).To(BeNil())

				totalRecords := 0
				for _, record := range r[0].Records() {
					record = record["kubernetes"].(map[string]interface{})["labels"].(map[string]interface{})
					Expect(record["name"].(string) == "app1").To(BeTrue())
					Expect(record["env"].(string) == "env1").To(BeTrue())
					totalRecords++
				}
				Expect(totalRecords >= 10).To(BeTrue())
				Expect(totalRecords <= 15).To(BeTrue())

				r, err = l.Query(fmt.Sprintf(LokiNsQuery, AllLogs), "", 40)
				Expect(err).To(BeNil())
				Expect(len(r[0].Records()) >= 10).To(BeTrue())
				Expect(len(r[0].Records()) <= 15).To(BeTrue())

			})

			It("applying ignore policy", func() {
				if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
					Skip("Skipping test since flow-control is not supported with fluentd")
				}

				f.Forwarder.Spec.Inputs[0].Application.GroupLimit = &logging.LimitSpec{
					Policy:              logging.DropPolicy,
					MaxRecordsPerSecond: 0,
				}

				f.Forwarder.Spec.Pipelines[0].InputRefs = append(f.Forwarder.Spec.Pipelines[0].InputRefs, logging.InputNameAudit)

				Expect(f.Deploy()).To(BeNil())
				Expect(WriteApplicationLogs(f, 1000)).To(Succeed())

				// shouldn't receive any logs from vector
				r, err := l.Query(fmt.Sprintf(LokiNsQuery, AllLogs), "", 20)
				Expect(err).To(BeNil())
				Expect(len(r)).To(Equal(0))
			})

		})

	})
})

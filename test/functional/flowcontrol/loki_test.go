// go:build !fluentd
package flowcontrol

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
)

var _ = Describe("[Functional][FlowControl] Policies at Output", func() {
	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		// Start a Loki server
		l = loki.NewReceiver(f.Namespace, "loki-server")
		Expect(l.Create(f.Test.Client)).To(Succeed())

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
				InputRefs:  []string{logging.InputNameApplication},
				OutputRefs: []string{logging.OutputTypeLoki},
				Name:       "flow-control",
			},
		)
	})

	AfterEach(func() {
		f.Cleanup()
	})

	Describe("when Output is Loki", func() {
		It("using flow control policy", func() {
			if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
				Skip("Skipping test since flow-control is not supported with fluentd")
			}

			f.Forwarder.Spec.Outputs[0].Limit = &logging.LimitSpec{
				MaxRecordsPerSecond: 10,
			}

			Expect(f.Deploy()).To(BeNil())
			Expect(f.WritesApplicationLogsWithDelay(1000, 0.0001)).To(Succeed())

			if _, err := l.QueryUntil(fmt.Sprintf(LokiNsQuery, AllLogs), "", 10); err != nil {
				Fail(fmt.Sprintf("Failed to read logs from Loki Server: %v", err))
			}
			r, err := l.Query(fmt.Sprintf(LokiNsQuery, AllLogs), "", 20)
			Expect(err).To(BeNil())
			records := r[0].Records()
			Expect(len(records) >= 10).To(BeTrue())
			Expect(len(records) <= 15).To(BeTrue())

		})
	})
})

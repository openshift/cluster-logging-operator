package flowcontrol

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
)

var _ = Describe("[Functional][FlowControl] Policies at Output", func() {
	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFramework()
		// Start a Loki server
		l = loki.NewReceiver(f.Namespace, "loki-server")
		Expect(l.Create(f.Test.Client)).To(Succeed())

		// Set up the common template forwarder configuration.
		obsruntime.NewClusterLogForwarderBuilder(f.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToLokiOutput(*l.InternalURL(""), func(spec *obs.OutputSpec) {
				spec.Limit = &obs.LimitSpec{
					MaxRecordsPerSecond: 10,
				}
			})
	})

	AfterEach(func() {
		f.Cleanup()
	})

	Context("when Output is rate limited", func() {
		It("should apply the policy for all output records", func() {

			Expect(f.Deploy()).To(BeNil())
			Expect(f.WritesApplicationLogsWithDelay(1000, 0.0001)).To(Succeed())

			if _, err := l.QueryUntil(fmt.Sprintf(LokiNsQuery, AllLogs), "", 10); err != nil {
				Fail(fmt.Sprintf("Failed to read logs from Loki Server: %v", err))
			}
			r, err := l.Query(fmt.Sprintf(LokiNsQuery, AllLogs), "", 20)
			Expect(err).To(BeNil())
			records := r[0].Records()
			Expect(len(records) >= 10).To(BeTrue(), fmt.Sprintf("Expected number of records %d >= 10", len(records)))
			Expect(len(records) <= 15).To(BeTrue(), fmt.Sprintf("Expected number of records %d <= 15", len(records)))

		})
	})
})

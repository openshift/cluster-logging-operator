package flowcontrol

import (
	"fmt"
	"time"
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
			const maxRecordsPerSecond = 10
			Expect(f.Deploy()).To(BeNil())

			// As in the input policy test: the throttle discards records over
			// the budget rather than queueing them, so write continuously at
			// well above the configured rate while sampling the Loki record
			// count, and verify the delivery rate tracks the configuration.
			writeDone := make(chan error, 1)
			go func() {
				// ~33 writes/s, sustained past the sampling window
				writeDone <- f.WritesApplicationLogsWithDelay(600, 0.03)
			}()
			defer func() {
				Expect(<-writeDone).To(Succeed(), "Expected the log writer to finish")
			}()

			query := fmt.Sprintf(LokiNsQuery, AllLogs)
			if _, err := l.QueryUntil(query, "", 10); err != nil {
				Fail(fmt.Sprintf("Failed to read logs from Loki Server: %v", err))
			}

			rate := observedDeliveryRate(l, query, 10*time.Second)
			Expect(rate).To(BeNumerically(">=", 2),
				fmt.Sprintf("Expected delivery rate %.2f records/s to be at least roughly the configured %d records/s", rate, maxRecordsPerSecond))
			Expect(rate).To(BeNumerically("<=", 4*maxRecordsPerSecond),
				fmt.Sprintf("Expected delivery rate %.2f records/s to be throttled near the configured %d records/s", rate, maxRecordsPerSecond))
		})
	})
})

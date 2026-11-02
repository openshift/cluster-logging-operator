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
			Expect(f.WritesApplicationLogsWithDelay(1000, 0.0001)).To(Succeed())

			query := fmt.Sprintf(LokiNsQuery, AllLogs)
			if _, err := l.QueryUntil(query, "", 10); err != nil {
				Fail(fmt.Sprintf("Failed to read logs from Loki Server: %v", err))
			}

			// The throttle (a governor quota of 10 records per 1s window)
			// paces delivery of the written backlog, so the number of records
			// stored in Loki grows at the configured rate. Asserting on a raw
			// record count is inherently racy: it depends on how much time
			// elapses between the writes and the query. Instead, measure the
			// observed delivery rate by sampling the Loki record count over a
			// fixed window of wall-clock time and verify it tracks the
			// configuration. The Loki entry timestamps cannot be used for this:
			// they are the events' own write-time timestamps, which all fall
			// within the ~1s it takes to write the backlog.
			rate := observedDeliveryRate(l, query, 10*time.Second)
			Expect(rate).To(BeNumerically(">=", 2),
				fmt.Sprintf("Expected delivery rate %.2f records/s to be at least roughly the configured %d records/s", rate, maxRecordsPerSecond))
			Expect(rate).To(BeNumerically("<=", 4*maxRecordsPerSecond),
				fmt.Sprintf("Expected delivery rate %.2f records/s to be throttled near the configured %d records/s", rate, maxRecordsPerSecond))
		})
	})
})

package flowcontrol

import (
	"fmt"
	"time"

	obsruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	LokiNsQuery    = `{kubernetes_namespace_name=~"%s"}`
	LokiAuditQuery = `{log_type="audit"}`
	AllLogs        = `.+`
)

var _ = Describe("[Functional][FlowControl] Policies at Input", func() {
	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFramework()
		// Start a Loki server
		l = loki.NewReceiver(f.Namespace, "loki-server")
		Expect(l.Create(f.Test.Client)).To(Succeed())
		obsruntime.NewClusterLogForwarderBuilder(f.Forwarder).
			FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
				spec.Name = "custom-app"
				spec.Application.Tuning = &obs.ContainerInputTuningSpec{
					RateLimitPerContainer: &obs.LimitSpec{
						MaxRecordsPerSecond: 10,
					},
				}
			}).ToLokiOutput(*l.InternalURL(""))
	})

	AfterEach(func() {
		f.Cleanup()
	})
	Context("when Source is Application", func() {
		It("should apply policy at the container level", func() {
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

// observedDeliveryRate samples the total record count stored in Loki at the
// start and end of a fixed wall-clock window and returns the delivery rate in
// records/second computed from the delta. It waits until the first records have
// arrived so the window measures the throttled steady state rather than the
// initial discovery delay.
func observedDeliveryRate(l *loki.Receiver, query string, window time.Duration) float64 {
	count := func() int {
		// The test Loki enforces max_entries_limit=5000, so the query limit
		// must stay at or below it or every query fails with a 400.
		streams, err := l.Query(query, "", 5000)
		ExpectWithOffset(2, err).To(BeNil(), "Expected no errors querying Loki")
		n := 0
		for _, s := range streams {
			n += len(s.Values)
		}
		return n
	}

	initial := count()
	start := time.Now()
	for time.Since(start) < window {
		time.Sleep(500 * time.Millisecond)
	}
	final := count()
	elapsed := time.Since(start).Seconds()
	ExpectWithOffset(2, elapsed).To(BeNumerically(">", 0))
	return float64(final-initial) / elapsed
}

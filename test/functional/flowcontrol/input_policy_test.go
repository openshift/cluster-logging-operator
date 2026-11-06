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
			Expect(f.Deploy()).To(BeNil())
			Expect(f.WritesApplicationLogsWithDelay(1000, 0.0001)).To(Succeed())

			if _, err := l.QueryUntil(fmt.Sprintf(LokiNsQuery, AllLogs), "", 10); err != nil {
				Fail(fmt.Sprintf("Failed to read logs from Loki Server: %v", err))
			}
			// Wait until at least 10 logs have been received, then verify the policy
			// by observing the rate at which NEW records arrive rather than a
			// cumulative count: the collector may discover the log file while the
			// write phase is still running, so an absolute upper bound on the total
			// depends on file-discovery latency, not on the rate limit.
			totalRecords := func() int {
				r, err := l.Query(fmt.Sprintf(LokiNsQuery, AllLogs), "", 1000)
				Expect(err).To(BeNil())
				count := 0
				for _, stream := range r {
					count += len(stream.Values)
				}
				return count
			}
			initial := totalRecords()
			Expect(initial >= 10).To(BeTrue(), fmt.Sprintf("Expected number of records %d >= 10", initial))

			// The threshold is 10 records/sec: after a 2s pause, at most
			// ~20 new records (plus slack for poll timing) should have arrived.
			// Without the policy, all 1000 written records would be ingested
			// nearly at once.
			time.Sleep(2 * time.Second)
			final := totalRecords()
			Expect(final-initial <= 30).To(BeTrue(), fmt.Sprintf("Expected at most 30 new records in 2s, got %d", final-initial))
			Expect(final < 1000).To(BeTrue(), fmt.Sprintf("Expected throttling to still be in progress, got %d records", final))
		})
	})
})

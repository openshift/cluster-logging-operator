package flowcontrol

import (
	"fmt"
	obsruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"

	. "github.com/onsi/ginkgo"
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
			// Wait until atleast 10 logs have been received
			r, err := l.Query(fmt.Sprintf(LokiNsQuery, AllLogs), "", 20)
			Expect(err).To(BeNil())
			records := r[0].Records()
			Expect(len(records) >= 10).To(BeTrue(), fmt.Sprintf("Expected number of records %d >= 10", len(records)))
			Expect(len(records) <= 15).To(BeTrue(), fmt.Sprintf("Expected number of records %d <= 10", len(records)))
		})
	})
})

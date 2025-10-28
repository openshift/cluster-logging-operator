package loki

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[Functional][Outputs][Loki] Forwarding to Loki", func() {
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
		obstestruntime.NewClusterLogForwarderBuilder(f.Forwarder).
			FromInput(obs.InputTypeAudit).
			ToLokiOutput(*l.InternalURL(""))

		Expect(f.Deploy()).To(BeNil())
	})

	AfterEach(func() {
		f.Cleanup()
	})

	writeAndVerifyLogs := func(writeLogs func() error) {

		Expect(writeLogs()).To(Succeed())

		// Verify we can query by Loki labels
		query := fmt.Sprintf(`{log_type=%q, kubernetes_host=%q}`, "audit", f.Pod.Spec.NodeName)
		r, err := l.QueryUntil(query, "", 1)
		Expect(err).To(Succeed())
		records := r[0].Records()
		Expect(records).To(HaveCap(1), "Exp. the record to be ingested")

		expLabels := map[string]string{
			"kubernetes_host": f.Pod.Spec.NodeName,
			"log_type":        string(obs.InputTypeAudit),
		}
		// quick fix since unable to disable service_name discovery via functional test arguments
		expLabels["service_name"] = "unknown_service"

		actualLabels := r[0].Stream
		Expect(actualLabels).To(BeEquivalentTo(expLabels), "Exp. labels to be added to the log record")
	}

	Context("when writing Audit logs from different sources", func() {
		It("should ingest kubernetes audit records from different audit sources without error", func() {
			now := time.Now()
			nowCrio := functional.CRIOTime(now)
			openshiftAuditLogs := fmt.Sprintf(functional.OpenShiftAuditLogTemplate, nowCrio, nowCrio)
			earlier := now.Add(-1 * 30 * time.Minute)
			earlierLog := functional.NewKubeAuditLog(earlier)

			writeAndVerifyLogs(func() error {
				Expect(f.WriteMessagesToOpenshiftAuditLog(openshiftAuditLogs, 1)).To(Succeed())
				return f.WriteMessagesTok8sAuditLog(earlierLog, 1)
			})
		})
	})

	Context("when writing Audit logs", func() {
		It("should ingest linux audit records without error", func() {
			writeAndVerifyLogs(func() error { return f.WriteAuditHostLog(1) })
		})
		It("should ingest kubernetes audit records without error", func() {
			writeAndVerifyLogs(func() error { return f.WriteK8sAuditLog(1) })
		})
		It("should ingest openshift audit records without error", func() {
			writeAndVerifyLogs(func() error { return f.WriteOpenshiftAuditLog(1) })
		})
		It("should ingest OVN audit records without error", func() {
			writeAndVerifyLogs(func() error { return f.WriteOVNAuditLog(1) })
		})
	})
})

//go:build fluentd
// +build fluentd

package loki

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"time"
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
				OutputRefs: []string{logging.OutputTypeLoki},
				InputRefs:  []string{logging.InputNameAudit},
				Labels:     map[string]string{"logging": "logging-value"},
			})

		Expect(f.Deploy()).To(BeNil())
	})

	AfterEach(func() {
		f.Cleanup()
	})

	writeAndVerifyLogs := func(auditType string, writeLogs func() error) {

		Expect(writeLogs()).To(Succeed())

		// Verify we can query by Loki labels
		query := fmt.Sprintf(`{log_type=%q, tag=%q}`, "audit", auditType)
		r, err := l.QueryUntil(query, "", 1)
		Expect(err).To(Succeed())
		records := r[0].Records()
		Expect(records).To(HaveCap(1), "Exp. the record to be ingested")

		expLabels := map[string]string{
			"kubernetes_host": functional.FunctionalNodeName,
			"tag":             auditType,
			"log_type":        "audit",
		}
		actualLabels := r[0].Stream
		delete(actualLabels, "fluentd_thread") // Added by loki plugin.
		Expect(actualLabels).To(BeEquivalentTo(expLabels), "Exp. labels to be added to the log record")

	}

	Context("when writing Audit logs from different sources", func() {

		//https://issues.redhat.com/browse/LOG-1702
		It("should ingest kubernetes audit records from different audit sources without error", func() {

			now := time.Now()
			nowCrio := functional.CRIOTime(now)
			openshiftAuditLogs := fmt.Sprintf(functional.OpenShiftAuditLogTemplate, nowCrio, nowCrio)
			earlier := now.Add(-1 * 30 * time.Minute)
			earlierLog := functional.NewKubeAuditLog(earlier)

			writeAndVerifyLogs("k8s-audit.log", func() error {
				Expect(f.WriteMessagesToOpenshiftAuditLog(openshiftAuditLogs, 1)).To(Succeed())
				return f.WriteMessagesTok8sAuditLog(earlierLog, 1)
			})
		})
	})
	Context("when writing Audit logs", func() {

		It("should ingest linux audit records without error", func() {
			writeAndVerifyLogs("linux-audit.log", func() error { return f.WriteAuditHostLog(1) })
		})
		It("should ingest kubernetes audit records without error", func() {
			writeAndVerifyLogs("k8s-audit.log", func() error { return f.WriteK8sAuditLog(1) })
		})
		It("should ingest openshift audit records without error", func() {
			writeAndVerifyLogs("openshift-audit.log", func() error { return f.WriteOpenshiftAuditLog(1) })
		})
		It("should ingest OVN audit records without error", func() {
			writeAndVerifyLogs("ovn-audit.log", func() error { return f.WriteOVNAuditLog(1) })
		})

	})

})

package metrics

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	rbacv1 "k8s.io/api/rbac/v1"
)

var _ = Describe("[Functional][Metrics] Discarded source logs metrics", func() {

	var (
		framework            *functional.CollectorFunctionalFramework
		metricsReaderRole    *rbacv1.ClusterRole
		metricsReaderBinding *rbacv1.ClusterRoleBinding
		tokenReviewBinding   *rbacv1.ClusterRoleBinding
	)

	AfterEach(func() {
		if tokenReviewBinding != nil {
			_ = framework.Test.Delete(tokenReviewBinding)
		}
		if metricsReaderBinding != nil {
			_ = framework.Test.Delete(metricsReaderBinding)
		}
		if metricsReaderRole != nil {
			_ = framework.Test.Delete(metricsReaderRole)
		}
		framework.Cleanup()
	})

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeAudit).
			ToHttpOutput()

		framework.VisitConfig = func(conf string) string {
			return strings.ReplaceAll(conf, "max_line_bytes = 3145728", "max_line_bytes = 256")
		}

		roleName := fmt.Sprintf("%s-metrics-reader", framework.Name)
		metricsReaderRole = runtime.NewClusterRole(
			roleName,
			runtime.NewNonResourceURLPolicyRule([]string{"/metrics"}, []string{"get"}),
		)
		Expect(framework.Test.Create(metricsReaderRole)).To(Succeed())

		metricsReaderBinding = runtime.NewClusterRoleBinding(
			roleName,
			runtime.NewClusterRoleRef(roleName),
			runtime.NewServiceAccountSubject("default", framework.Namespace),
		)
		Expect(framework.Test.Create(metricsReaderBinding)).To(Succeed())

		tokenReviewBinding = runtime.NewClusterRoleBinding(
			fmt.Sprintf("%s-token-reviewer", framework.Name),
			runtime.NewClusterRoleRef("system:auth-delegator"),
			runtime.NewServiceAccountSubject("default", framework.Namespace),
		)
		Expect(framework.Test.Create(tokenReviewBinding)).To(Succeed())
	})

	It("should generate vector_component_discarded_events_total when source logs exceed max_line_bytes", func() {
		Expect(framework.Deploy()).To(BeNil())

		auditLogFile := "/var/log/kube-apiserver/audit.log"

		// Write oversized lines (~1.5KB each, exceeding 256 byte limit) followed by a short line
		// in a single write so Vector processes them together in one read pass.
		longLine := functional.NewKubeAuditLog(time.Now())
		shortLine := `{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata"}`
		writeCmd := fmt.Sprintf(
			"mkdir -p %s && for i in $(seq 1 5); do echo '%s' >> %s; done && echo '%s' >> %s",
			"/var/log/kube-apiserver",
			strings.ReplaceAll(longLine, "'", "'\\''"),
			auditLogFile,
			shortLine,
			auditLogFile,
		)
		_, err := framework.RunCommand(constants.CollectorName, "bash", "-c", writeCmd)
		Expect(err).To(BeNil(), "failed to write audit log entries")

		metricsURL := fmt.Sprintf("https://%s.%s:24231/metrics", framework.Name, framework.Namespace)
		curlCmd := fmt.Sprintf(`curl -ks -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" %s`, metricsURL)
		grepDiscardCmd := fmt.Sprintf(`%s | grep -i discard`, curlCmd)

		Eventually(func() string {
			metrics, _ := framework.RunCommand(constants.CollectorName, "sh", "-c", grepDiscardCmd)
			return metrics
		}, 60*time.Second, 10*time.Second).Should(
			And(
				ContainSubstring("vector_component_discarded_events_total"),
				ContainSubstring(`component_kind="source"`),
			),
			"expected vector_component_discarded_events_total metric with component_kind=source",
		)
	})
})

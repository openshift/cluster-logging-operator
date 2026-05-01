package metrics

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	rbacv1 "k8s.io/api/rbac/v1"
)

var _ = Describe("[Functional][Metrics]Function testing of collector metrics", func() {

	const (
		sampleMetric = `# HELP vector_adaptive_concurrency_in_flight adaptive_concurrency_in_flight
# TYPE vector_adaptive_concurrency_in_flight histogram`
	)

	var (
		framework            *functional.CollectorFunctionalFramework
		metricsReaderRole    *rbacv1.ClusterRole
		metricsReaderBinding *rbacv1.ClusterRoleBinding
		tokenReviewBinding   *rbacv1.ClusterRoleBinding
	)

	AfterEach(func() {
		// Clean up cluster-scoped RBAC resources
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
			FromInput(obs.InputTypeApplication).
			ToHttpOutput()

		// Create ClusterRole to allow GET on /metrics
		roleName := fmt.Sprintf("%s-metrics-reader", framework.Name)
		metricsReaderRole = runtime.NewClusterRole(
			roleName,
			runtime.NewNonResourceURLPolicyRule([]string{"/metrics"}, []string{"get"}),
		)
		Expect(framework.Test.Create(metricsReaderRole)).To(Succeed())

		// Create ClusterRoleBinding to bind the service account to the metrics reader role
		metricsReaderBinding = runtime.NewClusterRoleBinding(
			roleName,
			runtime.NewClusterRoleRef(roleName),
			runtime.NewServiceAccountSubject("default", framework.Namespace),
		)
		Expect(framework.Test.Create(metricsReaderBinding)).To(Succeed())

		// Create ClusterRoleBinding to allow collector to do TokenReviews
		tokenReviewBinding = runtime.NewClusterRoleBinding(
			fmt.Sprintf("%s-token-reviewer", framework.Name),
			runtime.NewClusterRoleRef("system:auth-delegator"),
			runtime.NewServiceAccountSubject("default", framework.Namespace),
		)
		Expect(framework.Test.Create(tokenReviewBinding)).To(Succeed())
	})
	It("should return successfully when all outputs are up", func() {
		Expect(framework.Deploy()).To(BeNil())
		metricsURL := fmt.Sprintf("https://%s.%s:24231/metrics", framework.Name, framework.Namespace)
		curlCmd := fmt.Sprintf(`curl -ksv -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" %s`, metricsURL)
		metrics, _ := framework.RunCommand(constants.CollectorName, "sh", "-c", curlCmd)
		Expect(metrics).To(ContainSubstring(sampleMetric))
	})

	It("should return successfully even when the output is down", func() {
		Expect(framework.DeployWithVisitor(func(builder *runtime.PodBuilder) error { return nil })).To(BeNil())
		metricsURL := fmt.Sprintf("https://%s.%s:24231/metrics", framework.Name, framework.Namespace)
		curlCmd := fmt.Sprintf(`curl -ksv -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" %s`, metricsURL)
		metrics, _ := framework.RunCommand(constants.CollectorName, "sh", "-c", curlCmd)
		Expect(metrics).To(ContainSubstring(sampleMetric))
	})

})

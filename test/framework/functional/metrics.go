package functional

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
)

// SetupMetricsRBAC creates the cluster-scoped RBAC resources needed to scrape
// the collector's /metrics endpoint from within a test pod. Names are prefixed
// with the framework's namespace so parallel test packages cannot collide.
// The returned function deletes all created resources and should be registered
// with DeferCleanup or called in AfterEach.
func (f *CollectorFunctionalFramework) SetupMetricsRBAC() (metricsReaderRole *rbacv1.ClusterRole, metricsReaderBinding *rbacv1.ClusterRoleBinding, tokenReviewBinding *rbacv1.ClusterRoleBinding, err error) {
	roleName := fmt.Sprintf("%s-%s-metrics-reader", f.Test.NS.Name, f.Name)

	metricsReaderRole = runtime.NewClusterRole(
		roleName,
		runtime.NewNonResourceURLPolicyRule([]string{"/metrics"}, []string{"get"}),
	)
	if err = f.Test.Create(metricsReaderRole); err != nil {
		return nil, nil, nil, err
	}

	metricsReaderBinding = runtime.NewClusterRoleBinding(
		roleName,
		runtime.NewClusterRoleRef(roleName),
		runtime.NewServiceAccountSubject("default", f.Namespace),
	)
	if err = f.Test.Create(metricsReaderBinding); err != nil {
		return nil, nil, nil, err
	}

	tokenReviewName := fmt.Sprintf("%s-%s-token-reviewer", f.Test.NS.Name, f.Name)
	tokenReviewBinding = runtime.NewClusterRoleBinding(
		tokenReviewName,
		runtime.NewClusterRoleRef("system:auth-delegator"),
		runtime.NewServiceAccountSubject("default", f.Namespace),
	)
	if err = f.Test.Create(tokenReviewBinding); err != nil {
		return nil, nil, nil, err
	}

	return metricsReaderRole, metricsReaderBinding, tokenReviewBinding, nil
}

// CollectMetricLines polls the collector's Prometheus endpoint until a line
// matching both metricName and waitFor is found, then returns all lines
// matching metricName.
func (f *CollectorFunctionalFramework) CollectMetricLines(metricName, waitFor string, timeout time.Duration) ([]string, error) {
	var matched []string
	var lastErr error
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		raw, err := f.RunCommand(constants.CollectorName, "bash", "-c",
			fmt.Sprintf("curl -ks --max-time 10 -H \"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)\" https://%s.%s:24231/metrics", f.Name, f.Namespace))
		if err != nil {
			lastErr = err
			return false, nil
		}
		matched = nil
		for _, line := range strings.Split(raw, "\n") {
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.Contains(line, metricName) {
				matched = append(matched, line)
			}
		}
		for _, line := range matched {
			if strings.Contains(line, waitFor) {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil && lastErr != nil {
		return matched, fmt.Errorf("%w (last scrape error: %v)", err, lastErr)
	}
	return matched, err
}

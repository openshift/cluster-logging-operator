package loki

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

const lokiReceiverName = "loki-receiver"

func TestLogForwardingToLokiWithFluetnd(t *testing.T) {
	cl := runtime.NewClusterLogging()
	cl.Spec.Collection.Logs.Type = loggingv1.LogCollectionTypeFluentd
	clf := runtime.NewClusterLogForwarder()
	t.Run("default tenant", func(t *testing.T) { testDefaultTenant(t, cl, clf) })
	t.Run("custom tenant", func(t *testing.T) { testCustomTenant(t, cl, clf) })
}

func TestLogForwardingToLokiWithVector(t *testing.T) {
	cl := runtime.NewClusterLogging()
	cl.Spec.Collection.Logs.Type = loggingv1.LogCollectionTypeVector
	cl.Spec.Collection.Logs.FluentdSpec = loggingv1.FluentdSpec{}
	clf := runtime.NewClusterLogForwarder()
	t.Run("default tenant", func(t *testing.T) { testDefaultTenant(t, cl, clf) })
	t.Run("custom tenant", func(t *testing.T) { testCustomTenant(t, cl, clf) })
}

func setup(t *testing.T, c *client.Test, cl *loggingv1.ClusterLogging, clf *loggingv1.ClusterLogForwarder) (*loki.Receiver, *corev1.Pod) {
	rcv := loki.NewReceiver(c.NS.Name, "loki-receiver").EnableMultiTenant()
	gen := runtime.NewLogGenerator(c.NS.Name, rcv.Name, 100, 0, "I am Loki, of Asgard, and I am burdened with glorious purpose.")
	clf.Spec.Outputs[0].URL = rcv.InternalURL("").String()

	require.NoError(t, c.Recreate(cl))
	t.Cleanup(func() { _ = c.Delete(cl) })
	require.NoError(t, c.Recreate(clf))
	t.Cleanup(func() { _ = c.Delete(clf) })
	require.NoError(t, rcv.Create(c.Client))
	require.NoError(t, c.Create(gen))
	require.NoError(t, c.WaitFor(clf, client.ClusterLogForwarderReady))
	return rcv, gen
}

func testDefaultTenant(t *testing.T, cl *loggingv1.ClusterLogging, clf *loggingv1.ClusterLogForwarder) {
	clf.Spec = loggingv1.ClusterLogForwarderSpec{
		Outputs: []loggingv1.OutputSpec{{
			Name: lokiReceiverName,
			Type: loggingv1.OutputTypeLoki,
			// Default tenant, same as log_type: application, infrastructure, audit
		}},
		Pipelines: []loggingv1.PipelineSpec{
			{
				Name:       "test-app",
				InputRefs:  []string{loggingv1.InputNameApplication},
				OutputRefs: []string{lokiReceiverName},
			},
			{
				Name:       "test-audit",
				InputRefs:  []string{loggingv1.InputNameAudit},
				OutputRefs: []string{lokiReceiverName},
			},
			{
				Name:       "test-infra",
				InputRefs:  []string{loggingv1.InputNameInfrastructure},
				OutputRefs: []string{lokiReceiverName},
			},
		},
	}
	c := client.ForTest(t)
	rcv, _ := setup(t, c, cl, clf)
	for _, logType := range []string{"application", "infrastructure", "audit"} {
		logType := logType // Don't capture loop variable.
		t.Run(logType, func(t *testing.T) {
			r, err := rcv.QueryUntil(fmt.Sprintf(`{log_type=%q}`, logType), logType, 1)
			require.NoError(t, err, "failed query for %v", logType)
			require.Len(t, r, 1, "single log stream")
			assert.NotEmpty(t, r[0].Lines(), "no log lines for %v", logType)
			assert.Equal(t, logType, r[0].Stream["log_type"], "wrong type of logs")
		})
	}
}

func testCustomTenant(t *testing.T, cl *loggingv1.ClusterLogging, clf *loggingv1.ClusterLogForwarder) {
	clf.Spec = loggingv1.ClusterLogForwarderSpec{
		Outputs: []loggingv1.OutputSpec{{
			Name: lokiReceiverName,
			Type: loggingv1.OutputTypeLoki,
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Loki: &loggingv1.Loki{
					// Custom tenant key, pod_name
					TenantKey: "kubernetes.pod_name",
				},
			},
		}},
		Pipelines: []loggingv1.PipelineSpec{
			{
				Name:       "test-app",
				InputRefs:  []string{loggingv1.InputNameApplication},
				OutputRefs: []string{lokiReceiverName},
			},
		},
	}
	t.Cleanup(e2e.RunCleanupScript)
	c := client.ForTest(t)
	rcv, pod := setup(t, c, cl, clf)
	r, err := rcv.QueryUntil(`{log_type="application"}`, pod.Name, 1)
	require.NoError(t, err, "failed query for %v", pod.Name)
	require.Len(t, r, 1, "single log stream")
	assert.NotEmpty(t, r[0].Lines(), "no log lines for %v", pod.Name)
}

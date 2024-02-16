package loki

import (
	"fmt"
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

const lokiReceiverName = "loki-receiver"

var specs = []loggingv1.ClusterLogForwarderSpec{
	{
		Outputs: []loggingv1.OutputSpec{{
			Name: lokiReceiverName,
			Type: loggingv1.OutputTypeLoki,
		}},
		Pipelines: []loggingv1.PipelineSpec{
			{
				Name:       "test-app",
				InputRefs:  []string{loggingv1.InputNameApplication},
				OutputRefs: []string{lokiReceiverName},
				Labels:     map[string]string{"key1": "value1", "key2": "value2"},
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
	},
	{
		Outputs: []loggingv1.OutputSpec{{
			Name: lokiReceiverName,
			Type: loggingv1.OutputTypeLoki,
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Loki: &loggingv1.Loki{
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
	},
}

func TestLogForwardingToLokiWithFluentd(t *testing.T) {
	cl := runtime.NewClusterLogging()
	clf := runtime.NewClusterLogForwarder()
	for _, spec := range specs {
		clf.Spec = spec
		testLogForwardingToLoki(t, cl, clf)
	}
}

func TestLogForwardingToLokiWithVector(t *testing.T) {
	cl := runtime.NewClusterLogging()
	cl.Spec.Collection.Type = loggingv1.LogCollectionTypeVector
	cl.Spec.Collection.CollectorSpec = loggingv1.CollectorSpec{}
	clf := runtime.NewClusterLogForwarder()
	for _, spec := range specs {
		clf.Spec = spec
		testLogForwardingToLoki(t, cl, clf)
	}
}

func testLogForwardingToLoki(t *testing.T, cl *loggingv1.ClusterLogging, clf *loggingv1.ClusterLogForwarder) {
	c := client.ForTest(t)
	framework := e2e.NewE2ETestFramework()
	defer framework.Cleanup()
	framework.AddCleanup(func() error { return c.Delete(cl) })
	framework.AddCleanup(func() error { return c.Delete(clf) })

	rcv := loki.NewReceiver(c.NS.Name, "loki-receiver")
	framework.AddCleanup(func() error { return rcv.Delete(c.Client) })

	gen := runtime.NewLogGenerator(c.NS.Name, "log-generator", 100, 0, "I am Loki, of Asgard, and I am burdened with glorious purpose.")
	framework.AddCleanup(func() error { return c.Delete(gen) })
	clf.Spec.Outputs[0].URL = rcv.InternalURL("").String()

	// Start independent components in parallel to speed up the test.
	var g errgroup.Group
	e2e.RecreateClClfAsync(&g, c, cl, clf)
	g.Go(func() error { return rcv.Create(c.Client) })
	g.Go(func() error { return c.Create(gen) })

	require.NoError(t, g.Wait())
	require.NoError(t, c.WaitFor(clf, client.ClusterLogForwarderReady))
	require.NoError(t, framework.WaitFor(helpers.ComponentTypeCollector))

	// Now the actual test.
	for _, logType := range []string{"application", "infrastructure", "audit"} {
		r, err := rcv.QueryUntil(fmt.Sprintf(`{log_type=%q}`, logType), "", 1)
		require.NoError(t, err, "failed query for %v", logType)
		require.Len(t, r, 1, "single log stream")
		assert.NotEmpty(t, r[0].Lines(), "no log lines for %v", logType)
		assert.Equal(t, logType, r[0].Stream["log_type"], "wrong type of logs")
	}
}

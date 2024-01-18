package miscellaneous

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

const miscellaneousReceiverName = "miscellaneous-receiver"

var spec = loggingv1.ClusterLogForwarderSpec{
	Outputs: []loggingv1.OutputSpec{{
		Name: miscellaneousReceiverName,
		Type: loggingv1.OutputTypeLoki,
		URL:  "http://127.0.0.1:3100",
	}},
	Pipelines: []loggingv1.PipelineSpec{
		{
			Name:       "test-app",
			InputRefs:  []string{loggingv1.InputNameApplication},
			OutputRefs: []string{miscellaneousReceiverName},
			Labels:     map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			Name:       "test-audit",
			InputRefs:  []string{loggingv1.InputNameAudit},
			OutputRefs: []string{miscellaneousReceiverName},
		},
		{
			Name:       "test-infra",
			InputRefs:  []string{loggingv1.InputNameInfrastructure},
			OutputRefs: []string{miscellaneousReceiverName},
		},
	},
}

// TestLogForwardingWithEmptyCollection tests for issue https://github.com/openshift/cluster-logging-operator/issues/2312.
// It creates a CL with cl.Spec.Collection set to nil. This would trigger a nil pointer exception without a
// fix in place.
func TestLogForwardingWithEmptyCollection(t *testing.T) {
	// First, make sure that the Operator can handle a nil cl.Spec.Collection.
	// https://github.com/openshift/cluster-logging-operator/issues/2312
	t.Log("TestLogForwardingWithEmptyCollection: Test handling an empty ClusterLogging Spec.Condition")
	cl := runtime.NewClusterLogging()
	cl.Spec.Collection = nil
	clf := runtime.NewClusterLogForwarder()
	clf.Spec = spec

	c := client.ForTest(t)
	framework := e2e.NewE2ETestFramework()
	defer framework.Cleanup()
	framework.AddCleanup(func() error { return c.Delete(cl) })
	framework.AddCleanup(func() error { return c.Delete(clf) })
	var g errgroup.Group
	e2e.RecreateClClfAsync(&g, c, cl, clf)

	// We now expect to see a validation error.
	require.NoError(t, g.Wait())
	require.NoError(t, c.WaitFor(clf, client.ClusterLogForwarderValidationFailure))
	require.NoError(t, framework.WaitFor(helpers.ComponentTypeCollector))
}

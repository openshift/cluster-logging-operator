package miscellaneous

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

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

// TestLogForwardingWithEmptyCollection tests for issues https://github.com/openshift/cluster-logging-operator/issues/2312
// and https://github.com/openshift/cluster-logging-operator/issues/2314.
// It first creates a CL with cl.Spec.Collection set to nil. This would trigger a nil pointer exception without a
// fix in place.
// It then updates the CL to a valid status. Without a fix in place, the CLF's status would not update.
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

	// Now, make sure that the CLF's status updates to Ready when we update the CL resource to a valid status.
	// https://github.com/openshift/cluster-logging-operator/issues/2314
	t.Log("TestLogForwardingWithEmptyCollection: Make sure CLF updates when CL transitions to good state")
	clSpec := &loggingv1.CollectionSpec{
		Type:          loggingv1.LogCollectionTypeVector,
		CollectorSpec: loggingv1.CollectorSpec{},
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := c.Get(cl); err != nil {
			return err
		}
		cl.Spec.Collection = clSpec
		return c.Update(cl)
	})
	require.NoError(t, retryErr)
	// WaitFor alone will return too early and return an error. Instead, make use of the K8s retry framework and retry
	// up to 30 seconds.
	retryErr = retry.OnError(
		wait.Backoff{Steps: 10, Duration: 3 * time.Second, Factor: 1.0},
		func(error) bool { return true },
		func() error { t.Log("Retrieving CLF status"); return c.WaitFor(clf, client.ClusterLogForwarderReady) },
	)
	require.NoError(t, retryErr)
	require.NoError(t, framework.WaitFor(helpers.ComponentTypeCollector))
}

// TestLogForwardingWithEmptyCollection tests for issue https://github.com/openshift/cluster-logging-operator/issues/2315.
// Note that this issue can only be reproduced reliably with `make run` and does not seem to happen once the operator
// is deployed inside a pod - or respectively is more difficult to reproduce.
// The goal of the test is to make sure that the CLF resource remains stable for a sufficiently long duration - which
// we establish here as 15 seconds which in testing was enough to detect the issue.
func TestLogForwardingReconciliation(t *testing.T) {
	t.Log("TestLogForwardingWithEmptyCollection: Test handling an empty ClusterLogging Spec.Condition")
	cl := runtime.NewClusterLogging()
	clf := runtime.NewClusterLogForwarder()
	clf.Spec = spec

	c := client.ForTest(t)
	framework := e2e.NewE2ETestFramework()
	defer framework.Cleanup()
	framework.AddCleanup(func() error { return c.Delete(cl) })
	framework.AddCleanup(func() error { return c.Delete(clf) })
	var g errgroup.Group
	e2e.RecreateClClfAsync(&g, c, cl, clf)

	// We now expect to see no validation error.
	require.NoError(t, g.Wait())
	require.NoError(t, c.WaitFor(clf, client.ClusterLogForwarderReady))
	require.NoError(t, framework.WaitFor(helpers.ComponentTypeCollector))

	// Now, make sure that the CLF resource version remains unchanged. We run the test 5 times with 1 second between
	// the tests. The test itself lasts 15 seconds.
	retryErr := retry.OnError(
		wait.Backoff{Steps: 5, Duration: 1 * time.Second, Factor: 1.0},
		func(error) bool { return true },
		func() error {
			var resourceVersion string
			t.Log("Retrieving CLF status for the first time")
			if err := c.Get(clf); err != nil {
				return err
			}
			resourceVersion = clf.ResourceVersion
			t.Log("Sleeping for some time")
			time.Sleep(15 * time.Second)
			t.Log("Retrieving CLF status for the second time")
			if err := c.Get(clf); err != nil {
				return err
			}
			if resourceVersion != clf.ResourceVersion {
				t.Log("ResourceVersions do not match, CLF was updated. Retrying ...")
				return fmt.Errorf("ResourceVersion not stable, it changed from %q to %q",
					resourceVersion, clf.ResourceVersion)
			}
			return nil
		},
	)
	require.NoError(t, retryErr)
}

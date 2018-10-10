package e2e

import (
	"testing"
	//"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"time"

	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval = time.Second * 5
	timeout       = time.Second * 30
)

func TestClusterLogging(t *testing.T) {
	clusterloggingList := &loggingv1alpha1.ClusterLoggingList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: "logging.openshift.io/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(loggingv1alpha1.AddToScheme, clusterloggingList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("clusterlogging-group", func(t *testing.T) {
		t.Run("Cluster", ClusterLoggingCluster)
	})
}

func ClusterLoggingCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	err := ctx.InitializeClusterResources()
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Found namespace: %v", namespace)

	/* FIXME: this section won't pass until we have an image available so that
	   deploy/operator.yaml can deploy successfully
	*/
	// get global framework variables
	/*f := framework.Global
	// wait for cluster-logging-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "cluster-logging-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}*/
}

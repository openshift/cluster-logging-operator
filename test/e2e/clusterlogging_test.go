package e2e

import (
	goctx "context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"testing"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestClusterLogging(t *testing.T) {
	clusterloggingList := &logging.ClusterLoggingList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: logging.SchemeGroupVersion.String(),
		},
	}
	err := framework.AddToFrameworkScheme(logging.AddToScheme, clusterloggingList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("clusterlogging-group", func(t *testing.T) {
		t.Run("Cluster", ClusterLoggingCluster)
	})
}

func createRequiredSecret(f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	masterCASecret := utils.Secret(
		"logging-master-ca",
		namespace,
		map[string][]byte{
			"masterca":   utils.GetFileContents("test/files/ca.crt"),
			"masterkey":  utils.GetFileContents("test/files/ca.key"),
			"kibanacert": utils.GetFileContents("test/files/kibana-internal.crt"),
			"kibanakey":  utils.GetFileContents("test/files/kibana-internal.key"),
		},
	)

	err = f.Client.Create(goctx.TODO(), masterCASecret, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	return nil
}

func createRequiredClusterRoleAndBinding(f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	clusterRole := &rbac.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-logging-operator-priority",
		},
		Rules: []rbac.PolicyRule{
			rbac.PolicyRule{
				APIGroups: []string{"scheduling.k8s.io"},
				Resources: []string{"priorityclasses"},
				Verbs:     []string{"*"},
			},
		},
	}

	err = f.Client.Create(goctx.TODO(), clusterRole, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	clusterRoleBinding := &rbac.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-logging-operator-priority-rolebinding",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				Kind:      "ServiceAccount",
				Name:      "cluster-logging-operator",
				Namespace: namespace,
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-logging-operator-priority",
		},
	}

	err = f.Client.Create(goctx.TODO(), clusterRoleBinding, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func clusterLoggingFullClusterTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	// create clusterlogging custom resource
	exampleClusterLogging := &logging.ClusterLogging{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: logging.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-cluster-logging",
			Namespace: namespace,
		},
		Spec: logging.ClusterLoggingSpec{
			Visualization: logging.VisualizationSpec{
				Type: logging.VisualizationTypeKibana,
				KibanaSpec: logging.KibanaSpec{
					Replicas: 1,
				},
			},
			Curation: logging.CurationSpec{
				Type: logging.CurationTypeCurator,
				CuratorSpec: logging.CuratorSpec{
					Schedule: "* * * * *",
				},
			},
			Collection: logging.CollectionSpec{
				LogCollection: logging.LogCollectionSpec{
					Type: logging.LogCollectionTypeFluentd,
					FluentdSpec: logging.FluentdSpec{
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/infra": "true",
						},
					},
				},
			},
		},
	}
	err = f.Client.Create(goctx.TODO(), exampleClusterLogging, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "kibana-app", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "kibana-infra", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForCronJob(t, f.KubeClient, namespace, "curator-app", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForCronJob(t, f.KubeClient, namespace, "curator-infra", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForDaemonSet(t, f.KubeClient, namespace, "fluentd", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	return nil
}

func ClusterLoggingCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Found namespace: %v", namespace)

	// get global framework variables
	f := framework.Global
	// wait for cluster-logging-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "cluster-logging-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = createRequiredSecret(f, ctx); err != nil {
		t.Fatal(err)
	}

	if err = createRequiredClusterRoleAndBinding(f, ctx); err != nil {
		t.Fatal(err)
	}

	if err = clusterLoggingFullClusterTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

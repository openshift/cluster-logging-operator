package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 120
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

	elasticsearchList := &v1alpha1.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
	}
	err = framework.AddToFrameworkScheme(v1alpha1.AddToScheme, elasticsearchList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// run subtests
	t.Run("clusterlogging-group", func(t *testing.T) {

		t.Run("Cluster with fluentd", ClusterLoggingClusterFluentd)
		time.Sleep(time.Minute * 1) // wait for objects to be deleted/cleaned up
		t.Run("Cluster with rsyslog", ClusterLoggingClusterRsyslog)
	})
}

func clusterLoggingFullClusterTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, collector string) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	var collectionSpec logging.CollectionSpec
	if collector == "fluentd" {
		collectionSpec = logging.CollectionSpec{
			Logs: logging.LogCollectionSpec{
				Type:        logging.LogCollectionTypeFluentd,
				FluentdSpec: logging.FluentdSpec{},
			},
		}
	}
	if collector == "rsyslog" {
		collectionSpec = logging.CollectionSpec{
			Logs: logging.LogCollectionSpec{
				Type:        logging.LogCollectionTypeRsyslog,
				RsyslogSpec: logging.RsyslogSpec{},
			},
		}
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
			LogStore: logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 1,
				},
			},
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
			Collection:      collectionSpec,
			ManagementState: logging.ManagementStateManaged,
		},
	}
	err = f.Client.Create(goctx.TODO(), exampleClusterLogging, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "kibana", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForCronJob(t, f.KubeClient, namespace, "curator", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForDaemonSet(t, f.KubeClient, namespace, collector, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = clusterLoggingUpgradeClusterTest(t, f, ctx, collector)
	if err != nil {
		return err
	}

	return nil
}

func waitForOperatorToBeReady(t *testing.T, ctx *framework.TestCtx) error {
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return err
	}
	t.Logf("Found namespace: %v", namespace)

	// get global framework variables
	f := framework.Global
	// wait for cluster-logging-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "cluster-logging-operator", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	return nil
}

func ClusterLoggingClusterFluentd(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := waitForOperatorToBeReady(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err = clusterLoggingFullClusterTest(t, framework.Global, ctx, "fluentd"); err != nil {
		t.Fatal(err)
	}
}

func ClusterLoggingClusterRsyslog(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := waitForOperatorToBeReady(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err = clusterLoggingFullClusterTest(t, framework.Global, ctx, "rsyslog"); err != nil {
		t.Fatal(err)
	}
}

func clusterLoggingUpgradeClusterTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, collector string) error {

	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	currentOperator, err := f.KubeClient.AppsV1().Deployments(namespace).Get("cluster-logging-operator", metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return fmt.Errorf("failed to get currentOperator: %v", err)
	}

	currentEnv := currentOperator.Spec.Template.Spec.Containers[0].Env
	newEnv := []v1.EnvVar{
		{Name: "WATCH_NAMESPACE", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		{Name: "OPERATOR_NAME", Value: "cluster-logging-operator"},
		{Name: "ELASTICSEARCH_IMAGE", Value: "quay.io/openshift/origin-logging-elasticsearch5:upgraded"},
		{Name: "FLUENTD_IMAGE", Value: "quay.io/openshift/origin-logging-fluentd:upgraded"},
		{Name: "KIBANA_IMAGE", Value: "quay.io/openshift/origin-logging-kibana5:upgraded"},
		{Name: "CURATOR_IMAGE", Value: "quay.io/openshift/origin-logging-curator5:upgraded"},
		{Name: "OAUTH_PROXY_IMAGE", Value: "quay.io/openshift/origin-oauth-proxy:latest"},
		{Name: "RSYSLOG_IMAGE", Value: "quay.io/viaq/rsyslog:upgraded"},
	}

	currentOperator.Spec.Template.Spec.Containers[0].Env = newEnv
	err = f.Client.Update(goctx.TODO(), currentOperator)
	if err != nil {
		return fmt.Errorf("could not update cluster-logging-operator with updated image values %v", err)
	}

	err = CheckForElasticsearchImageName(t, f.Client, namespace, "elasticsearch", "quay.io/openshift/origin-logging-elasticsearch5:upgraded", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = CheckForDeploymentImageName(t, f.KubeClient, namespace, "kibana", "quay.io/openshift/origin-logging-kibana5:upgraded", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = CheckForCronJobImageName(t, f.KubeClient, namespace, "curator", "quay.io/openshift/origin-logging-curator5:upgraded", retryInterval, timeout)
	if err != nil {
		return err
	}

	if collector == "rsyslog" {
		err = CheckForDaemonSetImageName(t, f.KubeClient, namespace, collector, "quay.io/viaq/rsyslog:upgraded", retryInterval, timeout)
		if err != nil {
			return err
		}
	}
	if collector == "fluentd" {
		err = CheckForDaemonSetImageName(t, f.KubeClient, namespace, collector, "quay.io/openshift/origin-logging-fluentd:upgraded", retryInterval, timeout)
		if err != nil {
			return err
		}
	}

	currentOperator, err = f.KubeClient.AppsV1().Deployments(namespace).Get("cluster-logging-operator", metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return fmt.Errorf("failed to get currentOperator: %v", err)
	}

	currentOperator.Spec.Template.Spec.Containers[0].Env = currentEnv
	err = f.Client.Update(goctx.TODO(), currentOperator)
	if err != nil {
		return fmt.Errorf("could not update cluster-logging-operator with prior image values %v", err)
	}

	err = CheckForElasticsearchImageName(t, f.Client, namespace, "elasticsearch", "quay.io/openshift/origin-logging-elasticsearch5:latest", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = CheckForDeploymentImageName(t, f.KubeClient, namespace, "kibana", "quay.io/openshift/origin-logging-kibana5:latest", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = CheckForCronJobImageName(t, f.KubeClient, namespace, "curator", "quay.io/openshift/origin-logging-curator5:latest", retryInterval, timeout)
	if err != nil {
		return err
	}

	if collector == "rsyslog" {
		err = CheckForDaemonSetImageName(t, f.KubeClient, namespace, collector, "quay.io/viaq/rsyslog:latest", retryInterval, timeout)
		if err != nil {
			return err
		}
	}
	if collector == "fluentd" {
		err = CheckForDaemonSetImageName(t, f.KubeClient, namespace, collector, "quay.io/openshift/origin-logging-fluentd:latest", retryInterval, timeout)
		if err != nil {
			return err
		}
	}

	return nil
}

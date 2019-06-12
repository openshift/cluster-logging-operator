package e2e

import (
	goctx "context"
	"fmt"
	"strings"
	"testing"
	"time"

	elasticsearchapi "github.com/openshift/elasticsearch-operator/pkg/apis"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"

	loggingapi "github.com/openshift/cluster-logging-operator/pkg/apis"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	err := framework.AddToFrameworkScheme(loggingapi.AddToScheme, clusterloggingList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	elasticsearchList := &elasticsearch.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.SchemeGroupVersion.String(),
		},
	}
	err = framework.AddToFrameworkScheme(elasticsearchapi.AddToScheme, elasticsearchList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// run subtests
	t.Run("collectors", func(t *testing.T) {

		for _, collector := range []string{"fluentd", "rsyslog"} {
			t.Run(collector, func(t *testing.T) {
				ctx := framework.NewTestCtx(t)
				defer ctx.Cleanup()
				err := waitForOperatorToBeReady(t, ctx)
				if err != nil {
					t.Fatal(err)
				}

				if err = clusterLoggingInitialDeploymentTest(t, framework.Global, ctx, collector); err != nil {
					t.Fatal(err)
				}

				if err = clusterLoggingUpgradeTest(t, framework.Global, ctx, collector); err != nil {
					t.Fatal(err)
				}

				if err = changeLoggingCollectorTest(t, framework.Global, ctx, collector); err != nil {
					t.Fatal(err)
				}
			})
			time.Sleep(time.Minute * 1) // wait for objects to be deleted/cleaned up
		}
	})
}

func clusterLoggingInitialDeploymentTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, collector string) error {
	t.Log("Starting ClusterLogging initial deployment test...")
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

	// good default values for aws testing environment
	esResources := &v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("4Gi"),
			v1.ResourceCPU:    resource.MustParse("500m"),
		},
	}

	// create clusterlogging custom resource
	exampleClusterLogging := &logging.ClusterLogging{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: logging.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: namespace,
		},
		Spec: logging.ClusterLoggingSpec{
			LogStore: logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 1,
					Resources: esResources,
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
	t.Log("Completed ClusterLogging initial deployment test")

	return nil
}

func waitForOperatorToBeReady(t *testing.T, ctx *framework.TestCtx) error {
	t.Log("Initializing cluster resources...")
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
	t.Log("Waiting for cluster-logging-operator to be ready...")
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "cluster-logging-operator", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	return nil
}

func clusterLoggingUpgradeTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, collector string) error {
	t.Log("Starting ClusterLogging upgrade test...")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	currentOperator, err := f.KubeClient.AppsV1().Deployments(namespace).Get("cluster-logging-operator", metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return fmt.Errorf("failed to get currentOperator: %v", err)
	}

	newEnv := []v1.EnvVar{
		{Name: "WATCH_NAMESPACE", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		{Name: "POD_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
		{Name: "OPERATOR_NAME", Value: "cluster-logging-operator"},
		{Name: "ELASTICSEARCH_IMAGE", Value: "quay.io/openshift/origin-logging-elasticsearch5:v4.0"},
		{Name: "FLUENTD_IMAGE", Value: "quay.io/openshift/origin-logging-fluentd:v4.0"},
		{Name: "KIBANA_IMAGE", Value: "quay.io/openshift/origin-logging-kibana5:v4.0"},
		{Name: "CURATOR_IMAGE", Value: "quay.io/openshift/origin-logging-curator5:v4.0"},
		{Name: "OAUTH_PROXY_IMAGE", Value: "quay.io/openshift/origin-oauth-proxy:v4.0"},
		{Name: "RSYSLOG_IMAGE", Value: "docker.io/viaq/rsyslog:latest"},
	}

	t.Logf("Modified image ENV variables to force upgrade: %q", newEnv)
	currentOperator.Spec.Template.Spec.Containers[0].Env = newEnv
	err = f.Client.Update(goctx.TODO(), currentOperator)
	if err != nil {
		return fmt.Errorf("could not update cluster-logging-operator with updated image values %v", err)
	}

	err = CheckForElasticsearchImageName(t, f.Client, namespace, "elasticsearch", getValueFromEnvVar(newEnv, "ELASTICSEARCH_IMAGE"), retryInterval, timeout)
	if err != nil {
		return err
	}

	err = CheckForDeploymentImageName(t, f.KubeClient, namespace, "kibana", getValueFromEnvVar(newEnv, "KIBANA_IMAGE"), retryInterval, timeout)
	if err != nil {
		return err
	}
	err = CheckForDeploymentImageName(t, f.KubeClient, namespace, "kibana", getValueFromEnvVar(newEnv, "OAUTH_PROXY_IMAGE"), retryInterval, timeout)
	if err != nil {
		return err
	}

	err = CheckForCronJobImageName(t, f.KubeClient, namespace, "curator", getValueFromEnvVar(newEnv, "CURATOR_IMAGE"), retryInterval, timeout)
	if err != nil {
		return err
	}
	envKeyName := strings.ToUpper(collector) + "_IMAGE"
	err = CheckForDaemonSetImageName(t, f.KubeClient, namespace, collector, getValueFromEnvVar(newEnv, envKeyName), retryInterval, timeout)
	if err != nil {
		return err
	}

	t.Log("Completed ClusterLogging upgrade test")
	return nil
}

func changeLoggingCollectorTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, collector string) error {
	t.Log("Starting change logging collector test...")
	t.Logf("Logging Collection type is: %s", collector)
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}
	newcollector := ""
	exampleClusterLogging := &logging.ClusterLogging{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: logging.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: namespace,
		},
	}

	exampleName := types.NamespacedName{Name: exampleClusterLogging.GetName(), Namespace: namespace}
	if err = f.Client.Get(goctx.TODO(), exampleName, exampleClusterLogging); err != nil {
		return fmt.Errorf("failed to get exampleClusterLogging: %v", err)
	}

	if exampleClusterLogging.Spec.Collection.Logs.Type == logging.LogCollectionTypeFluentd {
		exampleClusterLogging.Spec.Collection.Logs.Type = logging.LogCollectionTypeRsyslog
		newcollector = "rsyslog"
	} else if exampleClusterLogging.Spec.Collection.Logs.Type == logging.LogCollectionTypeRsyslog {
		exampleClusterLogging.Spec.Collection.Logs.Type = logging.LogCollectionTypeFluentd
		newcollector = "fluentd"
	}

	err = f.Client.Update(goctx.TODO(), exampleClusterLogging)
	if err != nil {
		return fmt.Errorf("could not update exampleClusterLogging collection type: %v", err)
	}

	err = WaitForDaemonSet(t, f.KubeClient, namespace, newcollector, retryInterval, timeout)
	if err != nil {
		return err
	}

	t.Log("Completed changing ClusterLogging collection type test")
	return nil
}

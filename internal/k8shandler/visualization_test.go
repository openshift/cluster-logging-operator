package k8shandler

import (
	"context"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/visualization/console"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestRemoveKibanaCR(t *testing.T) {
	_ = es.SchemeBuilder.AddToScheme(scheme.Scheme)

	kbn := &es.Kibana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kibana",
			Namespace: "openshift-logging",
		},
		Spec: es.KibanaSpec{
			ManagementState: es.ManagementStateManaged,
			Replicas:        1,
		},
	}

	clr := &ClusterLoggingRequest{
		Cluster: &logging.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "openshift-logging",
			},
			Spec: logging.ClusterLoggingSpec{
				Visualization: &logging.VisualizationSpec{
					Type:   logging.VisualizationTypeKibana,
					Kibana: &logging.KibanaSpec{},
				},
			},
		},
	}
	clr.Client = fake.NewFakeClient(kbn) //nolint

	if err := clr.removeKibana(); err != nil {
		t.Error(err)
	}
}

func TestConsolePluginIsCreatedAndDeleted(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	cr := &ClusterLoggingRequest{
		Cluster:        runtime.NewClusterLogging(),
		Client:         c,
		ClusterVersion: "4.10.0",
	}
	cl := cr.Cluster
	// Enable korrel8r preview
	cl.Annotations = map[string]string{constants.AnnotationPreviewKorrel8rConsole: constants.Enabled}

	cl.Spec = logging.ClusterLoggingSpec{
		LogStore: &logging.LogStoreSpec{
			Type:      logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{Name: "some-loki-stack"},
		},
		Visualization: &logging.VisualizationSpec{
			Type:   logging.VisualizationTypeOCPConsole,
			Kibana: &logging.KibanaSpec{},
		},
	}
	r := console.NewReconciler(c, console.NewConfig(cl, "some-loki-stack-gateway-http", constants.Korrel8rNamespace, constants.Korrel8rName, []string{}), nil)
	cp := &consolev1alpha1.ConsolePlugin{}

	deploymentNSName := types.NamespacedName{Name: r.Name, Namespace: r.Namespace()}
	cpDeployment := &appv1.Deployment{}

	t.Run("create", func(t *testing.T) {
		require.NoError(t, cr.CreateOrUpdateVisualization())

		require.NoError(t, c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp))
		require.Contains(t, cp.Labels, constants.LabelK8sCreatedBy)
		assert.Equal(t, r.CreatedBy(), cp.Labels[constants.LabelK8sCreatedBy])
		assert.Equal(t, r.LokiService, cp.Spec.Proxy[0].Service.Name)
		assert.Equal(t, r.Korrel8rName, cp.Spec.Proxy[1].Service.Name)

		// Check deployment nodeSelector/Tolerations
		require.NoError(t, c.Get(context.Background(), deploymentNSName, cpDeployment))
		assert.Len(t, cpDeployment.Spec.Template.Spec.NodeSelector, 0)
		assert.Len(t, cpDeployment.Spec.Template.Spec.Tolerations, 0)
	})

	t.Run("delete", func(t *testing.T) {
		cl.Spec.LogStore = nil // Spec no longer wants console
		require.NoError(t, cr.CreateOrUpdateVisualization())
		err := c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp)
		assert.True(t, errors.IsNotFound(err), "expected NotFound got: %v", err)
	})
}

func TestConsolePluginIsCreatedAndDeleted_WithoutLokiStack(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	cl := runtime.NewClusterLogging()
	cl.Annotations = map[string]string{
		constants.AnnotationOCPConsoleMigrationTarget: "some-loki-stack",
	}
	cr := &ClusterLoggingRequest{
		Cluster:        cl,
		Client:         c,
		ClusterVersion: "4.10.0",
	}

	cl.Spec = logging.ClusterLoggingSpec{
		Visualization: &logging.VisualizationSpec{
			Type: logging.VisualizationTypeOCPConsole,
		},
	}
	r := console.NewReconciler(c, console.NewConfig(cl, "some-loki-stack-gateway-http", constants.Korrel8rNamespace, constants.Korrel8rName, []string{}), nil)
	cp := &consolev1alpha1.ConsolePlugin{}

	t.Run("create", func(t *testing.T) {
		require.NoError(t, cr.CreateOrUpdateVisualization())

		require.NoError(t, c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp))
		require.Contains(t, cp.Labels, constants.LabelK8sCreatedBy)
		assert.Equal(t, r.CreatedBy(), cp.Labels[constants.LabelK8sCreatedBy])
		assert.Len(t, cp.Spec.Proxy, 1)
		assert.Equal(t, r.LokiService, cp.Spec.Proxy[0].Service.Name)
	})

	t.Run("delete", func(t *testing.T) {
		cl.Annotations = nil
		cl.Spec.LogStore = nil // Spec no longer wants console
		require.NoError(t, cr.CreateOrUpdateVisualization())
		err := c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp)
		assert.True(t, errors.IsNotFound(err), "expected NotFound got: %v", err)
	})
}

func TestConsolePluginIsCreatedAndDeleted_WithoutKorrel8rPreview(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	cr := &ClusterLoggingRequest{
		Cluster:        runtime.NewClusterLogging(),
		Client:         c,
		ClusterVersion: "4.10.0",
	}
	cl := cr.Cluster
	// Don't enable korrel8r preview

	cl.Spec = logging.ClusterLoggingSpec{
		LogStore: &logging.LogStoreSpec{
			Type:      logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{Name: "some-loki-stack"},
		},
		Visualization: &logging.VisualizationSpec{
			Type:   logging.VisualizationTypeOCPConsole,
			Kibana: &logging.KibanaSpec{},
		},
	}
	r := console.NewReconciler(c, console.NewConfig(cl, "some-loki-stack-gateway-http", constants.Korrel8rNamespace, constants.Korrel8rName, []string{}), nil)
	cp := &consolev1alpha1.ConsolePlugin{}

	t.Run("create", func(t *testing.T) {
		require.NoError(t, cr.CreateOrUpdateVisualization())

		require.NoError(t, c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp))
		require.Contains(t, cp.Labels, constants.LabelK8sCreatedBy)
		assert.Equal(t, r.CreatedBy(), cp.Labels[constants.LabelK8sCreatedBy])
		assert.Len(t, cp.Spec.Proxy, 1)
		assert.Equal(t, r.LokiService, cp.Spec.Proxy[0].Service.Name)
	})

	t.Run("delete", func(t *testing.T) {
		cl.Spec.LogStore = nil // Spec no longer wants console
		require.NoError(t, cr.CreateOrUpdateVisualization())
		err := c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp)
		assert.True(t, errors.IsNotFound(err), "expected NotFound got: %v", err)
	})
}

func TestConsolePluginIsCreatedAndDeleted_WithCustomNodeSelectorAndToleration(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	cr := &ClusterLoggingRequest{
		Cluster:        runtime.NewClusterLogging(),
		Client:         c,
		ClusterVersion: "4.10.0",
	}
	cl := cr.Cluster
	// Don't enable korrel8r preview

	cl.Spec = logging.ClusterLoggingSpec{
		LogStore: &logging.LogStoreSpec{
			Type:      logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{Name: "some-loki-stack"},
		},
		Visualization: &logging.VisualizationSpec{
			Type: logging.VisualizationTypeOCPConsole,
			NodeSelector: map[string]string{
				"test": "test",
			},
			Tolerations: []v1.Toleration{
				{
					Key:   "test",
					Value: "test",
				},
			},
		},
	}
	r := console.NewReconciler(c, console.NewConfig(cl, "some-loki-stack-gateway-http", constants.Korrel8rNamespace, constants.Korrel8rName, []string{}), cl.Spec.Visualization)
	cp := &consolev1alpha1.ConsolePlugin{}

	deploymentNSName := types.NamespacedName{Name: r.Name, Namespace: r.Namespace()}
	cpDeployment := &appv1.Deployment{}

	t.Run("create", func(t *testing.T) {
		require.NoError(t, cr.CreateOrUpdateVisualization())

		require.NoError(t, c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp))
		require.Contains(t, cp.Labels, constants.LabelK8sCreatedBy)
		assert.Equal(t, r.CreatedBy(), cp.Labels[constants.LabelK8sCreatedBy])
		assert.Len(t, cp.Spec.Proxy, 1)
		assert.Equal(t, r.LokiService, cp.Spec.Proxy[0].Service.Name)

		// Check deployment nodeSelector/Tolerations
		require.NoError(t, c.Get(context.Background(), deploymentNSName, cpDeployment))
		assert.Len(t, cpDeployment.Spec.Template.Spec.NodeSelector, 1)
		assert.Len(t, cpDeployment.Spec.Template.Spec.Tolerations, 1)
		assert.Contains(t, cpDeployment.Spec.Template.Spec.NodeSelector, "test")
		assert.Contains(t, cpDeployment.Spec.Template.Spec.Tolerations, v1.Toleration{Key: "test", Value: "test"})
	})

	t.Run("delete", func(t *testing.T) {
		cl.Spec.LogStore = nil // Spec no longer wants console
		require.NoError(t, cr.CreateOrUpdateVisualization())
		err := c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp)
		assert.True(t, errors.IsNotFound(err), "expected NotFound got: %v", err)
	})
}

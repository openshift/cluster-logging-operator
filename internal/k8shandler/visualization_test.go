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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Cluster: runtime.NewClusterLogging(),
		Client:  c,
	}
	cl := cr.Cluster

	cl.Spec = logging.ClusterLoggingSpec{
		LogStore: &logging.LogStoreSpec{
			Type:      logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{Name: "some-loki-stack"},
		},
	}
	r := console.NewReconciler(c, console.NewConfig(cl, "some-loki-stack-gateway-http"))
	cp := &consolev1alpha1.ConsolePlugin{}

	t.Run("create", func(t *testing.T) {
		require.NoError(t, cr.CreateOrUpdateVisualization())

		require.NoError(t, c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp))
		require.Contains(t, cp.Labels, constants.LabelK8sCreatedBy)
		assert.Equal(t, r.CreatedBy(), cp.Labels[constants.LabelK8sCreatedBy])
		assert.Equal(t, r.LokiService, cp.Spec.Proxy[0].Service.Name)
	})

	t.Run("delete", func(t *testing.T) {
		cl.Spec.LogStore = nil // Spec no longer wants console
		require.NoError(t, cr.CreateOrUpdateVisualization())
		err := c.Get(context.Background(), client.ObjectKey{Name: r.Name}, cp)
		assert.True(t, errors.IsNotFound(err), "expected NotFound got: %v", err)
	})
}

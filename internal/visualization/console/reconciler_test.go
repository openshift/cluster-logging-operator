package console

import (
	"context"
	"testing"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func init() { utils.InitLogger("test") }

func fakeClient() client.Client { return fake.NewClientBuilder().WithScheme(scheme.Scheme).Build() }

var ctx = context.Background()

// assertConfig asserts that the Reconciler config matches the cluster state.
func assertConfig(t *testing.T, r *Reconciler) {
	c := r.c
	cp := &r.consolePlugin
	if assert.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(cp), cp)) {
		assert.Equal(t, r.Name, cp.Name)
		assert.Equal(t, r.Name, cp.Spec.Service.Name)

		assert.Equal(t, r.Namespace(), cp.Spec.Service.Namespace)
		assert.Equal(t, r.LokiService, cp.Spec.Proxy[0].Service.Name)
		assert.Equal(t, r.LokiPort, cp.Spec.Proxy[0].Service.Port)
		assert.Equal(t, r.CreatedBy(), cp.Labels[constants.LabelK8sCreatedBy])
	}

	d := &r.deployment
	if assert.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(d), d)) {
		assert.Equal(t, r.Image, d.Spec.Template.Spec.Containers[0].Image)
	}

	for _, o := range []client.Object{&r.deployment, &r.service, &r.configMap} {
		o := o
		kind := o.GetObjectKind().GroupVersionKind().Kind
		require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(o), o), kind)
		if assert.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(o), o), kind) {
			assert.Equal(t, r.Namespace(), o.GetNamespace(), kind)
			assert.Equal(t, r.CreatedBy(), o.GetLabels()[constants.LabelK8sCreatedBy])
			if assert.Len(t, o.GetOwnerReferences(), 1, kind) {
				oref := o.GetOwnerReferences()[0]
				assert.Equal(t, oref.Name, r.Owner.GetName(), kind)
				assert.Equal(t, oref.Kind, "ClusterLogging", kind)
				require.NotNil(t, oref.Controller, kind)
				assert.True(t, *oref.Controller, kind)
			}
		}

	}
}

// assertNotFound asserts that none of the objects for Config r exist.
func assertNotFound(t *testing.T, r *Reconciler) {
	c := r.c
	t.Helper()
	assert.NoError(t, r.each(func(m mutable) error {
		err := c.Get(ctx, client.ObjectKeyFromObject(m.o), m.o)
		assert.True(t, apierrors.IsNotFound(err), "expected not-found %v, got: %v", client.ObjectKeyFromObject(m.o), err)
		return nil
	}))
}

func TestVerifyResources(t *testing.T) {
	c := fakeClient()
	r := NewReconciler(c, NewConfig(runtime.NewClusterLogging(), "someservice", []string{}), nil)
	assert.NoError(t, r.Reconcile(ctx))
	assert.NoError(t, r.each(func(m mutable) error {
		kind := m.o.GetObjectKind().GroupVersionKind().String()
		t.Run("verify resource values "+kind, func(t *testing.T) {
			name := m.o.GetName()
			assert.Equal(t, "logging-view-plugin", name)
			assert.Equal(t, name, m.o.GetLabels()[constants.LabelApp])
			assert.Equal(t, name, m.o.GetLabels()[constants.LabelK8sName])
			assert.Equal(t, r.CreatedBy(), m.o.GetLabels()[constants.LabelK8sCreatedBy])
			if m.o.GetObjectKind().GroupVersionKind().Kind == "ConsolePlugin" {
				assert.Empty(t, m.o.GetNamespace())
				cp := &r.consolePlugin
				assert.Equal(t, consolev1alpha1.ConsolePluginService{
					Name:      name,
					Namespace: r.Namespace(),
					BasePath:  "/",
					Port:      r.pluginBackendPort(),
				}, cp.Spec.Service)
				assert.Equal(t, []consolev1alpha1.ConsolePluginProxy{{
					Type:      "Service",
					Alias:     "backend",
					Authorize: true,
					Service: consolev1alpha1.ConsolePluginProxyServiceConfig{
						Name:      "someservice",
						Namespace: "openshift-logging",
						Port:      8080,
					},
				}}, cp.Spec.Proxy)

			} else {
				assert.Equal(t, "openshift-logging", m.o.GetNamespace())
			}
		})
		return nil
	}))
}

func TestReconcileCreatesObjects(t *testing.T) {
	c := fakeClient()
	r := NewReconciler(c, NewConfig(runtime.NewClusterLogging(), "myLoki", []string{}), nil)
	require.NoError(t, r.Reconcile(ctx))
	assertConfig(t, r)
}

func TestReconcileUpdatesObjects(t *testing.T) {
	c := fakeClient()
	r := NewReconciler(c, NewConfig(runtime.NewClusterLogging(), "myLoki", []string{}), nil)
	require.NoError(t, r.Reconcile(ctx)) // Create objects
	assertConfig(t, r)

	// Modify configuration
	r.Image = "newimage"
	r.LokiService = "newloki"
	r.LokiPort = 42
	require.NoError(t, r.Reconcile(ctx)) // Create objects
	assertConfig(t, r)
}

func TestReconcilerDeletesObjects(t *testing.T) {
	c := fakeClient()
	r := NewReconciler(c, NewConfig(runtime.NewClusterLogging(), "myLoki", []string{}), nil)
	require.NoError(t, r.Reconcile(ctx)) // Create objects
	assertConfig(t, r)

	require.NoError(t, r.Delete(ctx))
	assertNotFound(t, r)
}

package k8shandler

import (
	"context"
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/consoleplugin"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	zeroInt = int32(0)
	twoInt  = int32(2)
)

func TestNewKibanaCR(t *testing.T) {
	tests := []struct {
		desc string
		cl   *logging.ClusterLogging
		want es.Kibana
	}{
		{
			desc: "default spec",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{},
					},
					LogStore: &logging.LogStoreSpec{
						Elasticsearch: &logging.ElasticsearchSpec{
							NodeCount: 1,
						},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        1,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
				},
			},
		},
		{
			desc: "no kibana replica no elasticsearch",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        0,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
				},
			},
		},
		{
			desc: "no kibana replica with elasticsearch",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{
							Replicas: &zeroInt,
						},
					},
					LogStore: &logging.LogStoreSpec{
						Elasticsearch: &logging.ElasticsearchSpec{
							NodeCount: 1,
						},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        0,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
				},
			},
		},
		{
			desc: "two kibana replica with elasticsearch",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{
							Replicas: &twoInt,
						},
					},
					LogStore: &logging.LogStoreSpec{
						Elasticsearch: &logging.ElasticsearchSpec{
							NodeCount: 1,
						},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        2,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
				},
			},
		},
		{
			desc: "custom resources",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					LogStore: &logging.LogStoreSpec{
						Elasticsearch: &logging.ElasticsearchSpec{
							NodeCount: 1,
						},
					},
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{
							Resources: &v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("136Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceMemory: defaultKibanaMemory,
									v1.ResourceCPU:    defaultKibanaCpuRequest,
								},
							},
							ProxySpec: logging.ProxySpec{
								Resources: &v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("136Mi"),
									},
									Requests: v1.ResourceList{
										v1.ResourceMemory: defaultKibanaMemory,
										v1.ResourceCPU:    defaultKibanaCpuRequest,
									},
								},
							},
						},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        1,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("136Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("136Mi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: defaultKibanaMemory,
								v1.ResourceCPU:    defaultKibanaCpuRequest,
							},
						},
					},
				},
			},
		},
		{
			desc: "custom node selectors",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					LogStore: &logging.LogStoreSpec{
						Elasticsearch: &logging.ElasticsearchSpec{
							NodeCount: 1,
						},
					},
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{
							NodeSelector: map[string]string{
								"test": "test",
							},
						},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        1,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
					NodeSelector: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			desc: "custom tolerations",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
					LogStore: &logging.LogStoreSpec{
						Elasticsearch: &logging.ElasticsearchSpec{
							NodeCount: 1,
						},
					},
					Visualization: &logging.VisualizationSpec{
						KibanaSpec: logging.KibanaSpec{
							Tolerations: []v1.Toleration{
								{
									Key:   "test",
									Value: "test",
								},
							},
						},
					},
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kibana",
					Namespace: "openshift-logging",
				},
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
					Replicas:        1,
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
					Tolerations: []v1.Toleration{
						{
							Key:   "test",
							Value: "test",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			got := newKibanaCustomResource(test.cl, "kibana")

			if got.Spec.ManagementState != test.want.Spec.ManagementState {
				t.Errorf("ManagementState: got %s, want %s", got.Spec.ManagementState, test.want.Spec.ManagementState)
			}

			if got.Spec.Replicas != test.want.Spec.Replicas {
				t.Errorf("%s, Replicas: got %d, want %d", test.desc, got.Spec.Replicas, test.want.Spec.Replicas)
			}

			if !reflect.DeepEqual(got.Spec.Resources, test.want.Spec.Resources) {
				t.Errorf("Resources: got\n%v\n\nwant\n%v", got.Spec.Resources, test.want.Spec.Resources)
			}

			if !reflect.DeepEqual(got.Spec.NodeSelector, test.want.Spec.NodeSelector) {
				t.Errorf("NodeSelector: got\n%v\n\nwant\n%v", got.Spec.NodeSelector, test.want.Spec.NodeSelector)
			}

			if !reflect.DeepEqual(got.Spec.Tolerations, test.want.Spec.Tolerations) {
				t.Errorf("Tolerations: got\n%v\n\nwant\n%v", got.Spec.Tolerations, test.want.Spec.Tolerations)
			}
		})
	}
}

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

	if err := clr.removeKibanaCR(); err != nil {
		t.Error(err)
	}
}

func TestIsKibanaCRDDifferent(t *testing.T) {
	tests := []struct {
		desc    string
		current *es.Kibana
		desired *es.Kibana
	}{
		{
			desc: "management state",
			current: &es.Kibana{
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateManaged,
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					ManagementState: es.ManagementStateUnmanaged,
				},
			},
		},
		{
			desc: "replicas",
			current: &es.Kibana{
				Spec: es.KibanaSpec{
					Replicas: 1,
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					Replicas: 2,
				},
			},
		},
		{
			desc: "node selectors",
			current: &es.Kibana{
				Spec: es.KibanaSpec{
					NodeSelector: map[string]string{
						"sel1": "value1",
					},
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					NodeSelector: map[string]string{
						"sel1": "value1",
						"sel2": "value2",
					},
				},
			},
		},
		{
			desc: "tolerations",
			current: &es.Kibana{
				Spec: es.KibanaSpec{
					Tolerations: []v1.Toleration{},
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					Tolerations: []v1.Toleration{
						{
							Key: "test",
						},
					},
				},
			},
		},
		{
			desc: "resources",
			current: &es.Kibana{
				Spec: es.KibanaSpec{
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
							v1.ResourceCPU:    defaultKibanaCpuRequest,
						},
					},
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: defaultKibanaMemory,
						},
					},
				},
			},
		},
		{
			desc: "proxy resources",
			current: &es.Kibana{
				Spec: es.KibanaSpec{
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: defaultKibanaMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: defaultKibanaMemory,
								v1.ResourceCPU:    defaultKibanaCpuRequest,
							},
						},
					},
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: defaultKibanaMemory,
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			got := isKibanaCRDDifferent(test.current, test.desired)
			if !got {
				t.Errorf("kibana crd not marked different, got %t", got)
			}
			if !reflect.DeepEqual(test.current.Spec, test.desired.Spec) {
				t.Errorf("kibana CR current spec not matching desired for %s, got %v, want %v", test.desc, test.current.Spec, test.desired.Spec)
			}
		})
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
	r := consoleplugin.NewReconciler(c, consoleplugin.NewConfig(cl, "some-loki-stack-gateway-http"))
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

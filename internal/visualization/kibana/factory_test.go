package kibana_test

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/visualization/kibana"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
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
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
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
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
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
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
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
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
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
									v1.ResourceMemory: kibana.DefaultKibanaMemory,
									v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
								},
							},
							ProxySpec: logging.ProxySpec{
								Resources: &v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("1986Mi"),
									},
									Requests: v1.ResourceList{
										v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
										v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
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
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
						},
					},
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("1986Mi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
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
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
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
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaProxyMemory,
								v1.ResourceCPU:    kibana.DefaultKibanaProxyCpuRequest,
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
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
			got := kibana.New(constants.OpenshiftNS, constants.KibanaName, test.cl.Spec.Visualization, test.cl.Spec.LogStore, utils.AsOwner(test.cl))

			if got.Spec.ManagementState != test.want.Spec.ManagementState {
				t.Errorf("ManagementState: got %s, want %s", got.Spec.ManagementState, test.want.Spec.ManagementState)
			}

			if got.Spec.Replicas != test.want.Spec.Replicas {
				t.Errorf("%s, Replicas: got %d, want %d", test.desc, got.Spec.Replicas, test.want.Spec.Replicas)
			}
			if !utils.AreResourcesSame(got.Spec.ProxySpec.Resources, test.want.Spec.ProxySpec.Resources) {
				t.Errorf("Proxy Resources: got\n%v\n\nwant\n%v", got.Spec.ProxySpec.Resources, test.want.Spec.ProxySpec.Resources)
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

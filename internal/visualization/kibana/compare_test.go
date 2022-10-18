package kibana_test

import (
	"github.com/openshift/cluster-logging-operator/internal/visualization/kibana"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestKibanaAreSame(t *testing.T) {
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
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
							v1.ResourceCPU:    kibana.DefaultKibanaCpuRequest,
						},
					},
				},
			},
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: kibana.DefaultKibanaMemory,
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
			desired: &es.Kibana{
				Spec: es.KibanaSpec{
					ProxySpec: es.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: kibana.DefaultKibanaMemory,
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
			got, _ := kibana.AreSame(*test.current, *test.desired)
			if got {
				t.Errorf("kibana cr not marked different, got %t", got)
			}
		})
	}
}

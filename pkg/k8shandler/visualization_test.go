package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	es "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//TODO: Remove this in the next release after removing old kibana code completely
func TestHasCLORef(t *testing.T) {
	clr := ClusterLoggingRequest{
		client: nil,
		cluster: &logging.ClusterLogging{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:                       "cluster-logging",
				GenerateName:               "",
				Namespace:                  "",
				SelfLink:                   "",
				UID:                        "123",
				ResourceVersion:            "",
				Generation:                 0,
				CreationTimestamp:          metav1.Time{},
				DeletionTimestamp:          nil,
				DeletionGracePeriodSeconds: nil,
				Labels:                     nil,
				Annotations:                nil,
				OwnerReferences:            nil,
				Initializers:               nil,
				Finalizers:                 nil,
				ClusterName:                "",
			},
			Spec:   logging.ClusterLoggingSpec{},
			Status: logging.ClusterLoggingStatus{},
		},
		ForwarderSpec: logging.ClusterLogForwarderSpec{},
		Collector:     nil,
	}

	obj := &apps.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "test-deployment",
			GenerateName:               "",
			Namespace:                  "",
			SelfLink:                   "",
			UID:                        "",
			ResourceVersion:            "",
			Generation:                 0,
			CreationTimestamp:          metav1.Time{},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Initializers:               nil,
			Finalizers:                 nil,
			ClusterName:                "",
		},
		Spec:   apps.DeploymentSpec{},
		Status: apps.DeploymentStatus{},
	}

	utils.AddOwnerRefToObject(obj, utils.AsOwner(clr.cluster))

	t.Log("refs:", obj.GetOwnerReferences())
	if !HasCLORef(obj, &clr) {
		t.Error("no owner reference found but it should be found")
	}
}

func TestAreRefsEqual(t *testing.T) {
	r1 := metav1.OwnerReference{
		APIVersion: logging.SchemeGroupVersion.String(),
		Kind:       "ClusterLogging",
		Name:       "testRef",
		Controller: func() *bool { t := true; return &t }(),
	}

	r2 := metav1.OwnerReference{
		APIVersion: logging.SchemeGroupVersion.String(),
		Kind:       "ClusterLogging",
		Name:       "testRef",
		Controller: func() *bool { t := true; return &t }(),
	}

	if !AreRefsEqual(&r1, &r2) {
		t.Error("refs are not equal but they should be")
	}
}

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
				},
			},
			want: es.Kibana{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Kibana",
					APIVersion: es.SchemeGroupVersion.String(),
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
			desc: "custom resources",
			cl: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: "openshift-logging",
				},
				Spec: logging.ClusterLoggingSpec{
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
					APIVersion: es.SchemeGroupVersion.String(),
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
					APIVersion: es.SchemeGroupVersion.String(),
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
					APIVersion: es.SchemeGroupVersion.String(),
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
				t.Errorf("Replicas: got %d, want %d", got.Spec.Replicas, test.want.Spec.Replicas)
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

package k8shandler

import (
	"fmt"
	"reflect"
	"testing"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGetReadinessProbe(t *testing.T) {
	goodProbe := v1.Probe{
		TimeoutSeconds:      30,
		InitialDelaySeconds: 10,
		FailureThreshold:    15,
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromInt(9300),
			},
		},
	}
	if !reflect.DeepEqual(goodProbe, getReadinessProbe()) {
		t.Errorf("Probe was incorrect: %v", getReadinessProbe())
	}
}

func TestGetAffinity(t *testing.T) {
	rolesArray := [][]string{{"master"}, {"client", "data", "master"},
		{"client", "data"}, {"data"}, {"client"}}
	goodAffinities := []v1.Affinity{}
	for _, roles := range rolesArray {
		labelSelectorReqs := []metav1.LabelSelectorRequirement{}
		for _, role := range roles {
			labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
				Key:      fmt.Sprintf("es-node-%s", role),
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"true"},
			})
		}
		aff := v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: labelSelectorReqs,
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}
		goodAffinities = append(goodAffinities, aff)
	}

	for i, roles := range rolesArray {
		ndRoles := []v1alpha1.ElasticsearchNodeRole{}
		for _, role := range roles {
			ndRoles = append(ndRoles, v1alpha1.ElasticsearchNodeRole(role))
		}
		cfg := desiredNodeState{
			Roles: ndRoles,
		}
		if !reflect.DeepEqual(goodAffinities[i], cfg.getAffinity()) {
			t.Errorf("Incorrect v1.Affinity constructed for role setb: %v", roles)

		}
	}
}

func TestGetResourceRequirements(t *testing.T) {
	CPU1, _ := resource.ParseQuantity("110m")
	CPU2, _ := resource.ParseQuantity("210m")
	Mem1, _ := resource.ParseQuantity("257Mi")
	Mem2, _ := resource.ParseQuantity("513Mi")
	defMemLim, _ := resource.ParseQuantity(defaultMemoryLimit)
	defCPUReq, _ := resource.ParseQuantity(defaultCPURequest)
	defMemReq, _ := resource.ParseQuantity(defaultMemoryRequest)
	defCPULim, _ := resource.ParseQuantity(defaultCPULimit)

	resList1 := v1.ResourceList{
		"cpu": CPU1,
	}
	resList2 := v1.ResourceList{
		"memory": Mem1,
	}
	resList3 := v1.ResourceList{
		"cpu":    CPU2,
		"memory": Mem2,
	}
	req1 := v1.ResourceRequirements{
		Limits:   resList1,
		Requests: resList2,
	}
	req2 := v1.ResourceRequirements{
		Limits:   resList2,
		Requests: resList1,
	}
	req3 := v1.ResourceRequirements{
		Limits:   resList3,
		Requests: resList3,
	}
	req4 := v1.ResourceRequirements{}
	resReq1 := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    CPU1,
			"memory": Mem1,
		},
		Requests: v1.ResourceList{
			"cpu":    CPU1,
			"memory": Mem1,
		},
	}
	resReq2 := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    CPU1,
			"memory": Mem2,
		},
		Requests: v1.ResourceList{
			"cpu":    CPU2,
			"memory": Mem1,
		},
	}
	resReq3 := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    CPU1,
			"memory": defMemLim,
		},
		Requests: v1.ResourceList{
			"cpu":    defCPUReq,
			"memory": Mem1,
		},
	}
	resReq4 := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    defCPULim,
			"memory": Mem1,
		},
		Requests: v1.ResourceList{
			"cpu":    CPU1,
			"memory": defMemReq,
		},
	}

	var table = []struct {
		commonRequirements v1.ResourceRequirements
		nodeRequirements   v1.ResourceRequirements
		result             v1.ResourceRequirements
	}{
		{req1, req2, resReq1},
		{req2, req1, resReq1},
		{req1, req3, req3},
		{req3, req1, resReq2},
		{req1, req4, resReq3},
		{req4, req1, resReq3},
		{req2, req4, resReq4},
	}

	for _, tt := range table {
		actual := getResourceRequirements(tt.commonRequirements, tt.nodeRequirements)
		if !reflect.DeepEqual(actual, tt.result) {
			t.Errorf("Incorrect v1.ResourceRequirements constructed: %v, should be: %v", actual, tt.result)

		}
	}
}

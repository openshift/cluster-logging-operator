package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewFluentdPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name")

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}

	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != defaultFluentdMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultFluentdMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultFluentdMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultFluentdMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultFluentdCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultFluentdCpuRequest)
	}

	CheckIfThereIsOnlyTheLinuxSelector(podSpec, t)
}

func TestNewFluentdPodSpecWhenResourcesAreDefined(t *testing.T) {
	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
					},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name")

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}

	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the spec memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the spec memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the spec CPU request to be %v", requestCPU)
	}
}

func TestFluentdPodSpecHasTaintTolerations(t *testing.T) {

	expectedTolerations := []v1.Toleration{
		v1.Toleration{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
		v1.Toleration{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name")

	if !reflect.DeepEqual(podSpec.Tolerations, expectedTolerations) {
		t.Errorf("Exp. the tolerations to be %q but was %q", expectedTolerations, podSpec.Tolerations)
	}
}

func TestNewFluentdPodSpecWhenSelectorIsDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo": "bar",
	}
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						NodeSelector: expSelector,
					},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name")

	if !reflect.DeepEqual(podSpec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, podSpec.NodeSelector)
	}
}

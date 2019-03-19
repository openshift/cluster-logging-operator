package k8shandler

import (
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewKibanaPodSpecWhenResourcesAreUndefined(t *testing.T) {

	cluster := &ClusterLogging{&logging.ClusterLogging{}}
	podSpec := cluster.newKibanaPodSpec("test-app-name", "elasticsearch")

	if len(podSpec.Containers) != 2 {
		t.Error("Exp. there to be 2 container")
	}

	//check kibana
	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != defaultKibanaMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultKibanaMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultKibanaMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultKibanaMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultKibanaCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultKibanaCpuRequest)
	}
	//check proxy
	resources = podSpec.Containers[1].Resources
	if resources.Limits[v1.ResourceMemory] != defaultKibanaProxyMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultKibanaMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultKibanaProxyMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultKibanaProxyMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultKibanaCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultKibanaCpuRequest)
	}
}

func TestNewKibanaPodSpecWhenResourcesAreDefined(t *testing.T) {
	cluster := &ClusterLogging{
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				Visualization: logging.VisualizationSpec{
					Type: "kibana",
					KibanaSpec: logging.KibanaSpec{
						Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
						ProxySpec: logging.ProxySpec{
							Resources: newResourceRequirements("200Gi", "", "220Gi", "2500m"),
						},
					},
				},
			},
		},
	}
	podSpec := cluster.newKibanaPodSpec("test-app-name", "elasticsearch")

	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")

	if len(podSpec.Containers) != 2 {
		t.Error("Exp. there to be 2 container")
	}

	//check kibana
	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the default memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the default memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the default CPU request to be %v", requestCPU)
	}

	limitMemory = resource.MustParse("200Gi")
	requestMemory = resource.MustParse("220Gi")
	requestCPU = resource.MustParse("2500m")
	//check proxy
	resources = podSpec.Containers[1].Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the default memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the default memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the default CPU request to be %v", requestCPU)
	}

}

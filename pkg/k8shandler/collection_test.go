package k8shandler

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewFluentdPodSpecWhenResourcesAreUndefined(t *testing.T) {

	cluster := &v1alpha1.ClusterLogging{}
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
}
func TestNewFluentdPodSpecWhenResourcesAreDefined(t *testing.T) {
	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	cluster := &v1alpha1.ClusterLogging{
		Spec: v1alpha1.ClusterLoggingSpec{
			Collection: v1alpha1.CollectionSpec{
				v1alpha1.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: v1alpha1.FluentdSpec{
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
func TestNewRsyslogPodSpecWhenResourcesAreUndefined(t *testing.T) {

	cluster := &v1alpha1.ClusterLogging{}
	podSpec := newRsyslogPodSpec(cluster, "test-app-name", "test-infra-name")

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}

	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != defaultRsyslogMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultRsyslogMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultRsyslogMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultRsyslogMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultFluentdCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultRsyslogCpuRequest)
	}
}
func TestNewRsyslogPodSpecWhenResourcesAreDefined(t *testing.T) {
	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	cluster := &v1alpha1.ClusterLogging{
		Spec: v1alpha1.ClusterLoggingSpec{
			Collection: v1alpha1.CollectionSpec{
				v1alpha1.LogCollectionSpec{
					Type: "rsyslog",
					RsyslogSpec: v1alpha1.RsyslogSpec{
						Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
					},
				},
			},
		},
	}
	podSpec := newRsyslogPodSpec(cluster, "test-app-name", "test-infra-name")

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

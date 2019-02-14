package k8shandler

import (
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewCuratorCronJobWhenResourcesAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	cronJob := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	podSpec := cronJob.Spec.JobTemplate.Spec.Template.Spec

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 container")
	}

	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != defaultCuratorMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultCuratorMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultCuratorMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultCuratorMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultFluentdCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultCuratorCpuRequest)
	}
}

func TestNewCuratorCronJobWhenResourcesAreDefined(t *testing.T) {
	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: logging.CurationSpec{
				Type: "curator",
				CuratorSpec: logging.CuratorSpec{
					Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
				},
			},
		},
	}
	cronJob := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	podSpec := cronJob.Spec.JobTemplate.Spec.Template.Spec

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

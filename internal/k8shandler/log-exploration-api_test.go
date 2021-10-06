package k8shandler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

var _ = Describe("log-exploration-api.go#newLogExplorationAPIPodSpec", func() {
	var (
		podSpec   v1.PodSpec
		container v1.Container
	)
	BeforeEach(func() {
		podSpec = newLogExplorationApiPodSpec()
		container = podSpec.Containers[0]
	})
	Describe("when creating the logexplorationapi container", func() {

		It("should provide the pod IP as an environment var", func() {
			Expect(container.Env).To(IncludeEnvVar(v1.EnvVar{Name: "POD_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						APIVersion: "v1", FieldPath: "status.podIP"}}}))
		})
	})

})

func TestNewLogExplorationAPIPodSpecWhenResourcesAreDefined(t *testing.T) {
	limitMemory := resource.MustParse("736Mi")
	limitCpu := resource.MustParse("100m")
	requestMemory := resource.MustParse("736Mi")
	requestCPU := resource.MustParse("100m")
	podSpec := newLogExplorationApiPodSpec()

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}
	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the spec memory limit to be %v", limitMemory)
	}
	if resources.Limits[v1.ResourceCPU] != limitCpu {
		t.Errorf("Exp. the spec memory limit to be %v", limitCpu)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the spec memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the spec CPU request to be %v", requestCPU)
	}
}

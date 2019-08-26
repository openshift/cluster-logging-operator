package k8shandler

import (
	"reflect"
	"strings"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewKibanaPodSpecSetsProxyToUseServiceAccountAsOAuthClient(t *testing.T) {
	clusterlogging := &logging.ClusterLogging{}
	spec := newKibanaPodSpec(clusterlogging, "kibana", "elasticsearch")
	valueIsCorrect := false
	for _, arg := range spec.Containers[1].Args {
		keyValue := strings.Split(arg, "=")
		if len(keyValue) >= 2 && keyValue[0] == "-client-id" {
			valueIsCorrect = keyValue[1] == "system:serviceaccount:openshift-logging:kibana"
		}
	}
	if !valueIsCorrect {
		t.Error("Exp. the proxy container arg 'client-id=system:serviceaccount:openshift-logging:kibana'")
	}
}

func TestNewKibanaPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch")

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
	// check node selecor
	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contains the linux allocation selector but was %T", podSpec.NodeSelector)
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
	cluster := &logging.ClusterLogging{
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
	}
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch")

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
func TestNewKibanaPodSpecWhenNodeSelectorIsDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo":             "bar",
		utils.OsNodeLabel: utils.LinuxValue,
	}
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Visualization: logging.VisualizationSpec{
				Type: "kibana",
				KibanaSpec: logging.KibanaSpec{
					NodeSelector: expSelector,
				},
			},
		},
	}
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch")

	//check kibana
	if !reflect.DeepEqual(podSpec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, podSpec.NodeSelector)
	}

}

func TestNewKibanaPodNoTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Visualization: logging.VisualizationSpec{
				Type:       "kibana",
				KibanaSpec: logging.KibanaSpec{},
			},
		},
	}

	podSpec := newKibanaPodSpec(cluster, "test-app-name", "test-infra-name")
	tolerations := podSpec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewKibanaPodWithTolerations(t *testing.T) {

	expTolerations := []v1.Toleration{
		v1.Toleration{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Visualization: logging.VisualizationSpec{
				Type: "kibana",
				KibanaSpec: logging.KibanaSpec{
					Tolerations: expTolerations,
				},
			},
		},
	}

	podSpec := newKibanaPodSpec(cluster, "test-app-name", "test-infra-name")
	tolerations := podSpec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

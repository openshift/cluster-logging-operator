package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateKibanaProxyClusterRoleBindingWithoutError(t *testing.T) {
	clusterRoleBinding := &rbac.ClusterRoleBinding{}

	clusterLoggingRequest := &ClusterLoggingRequest{
		client:  fake.NewFakeClient(clusterRoleBinding),
		cluster: &logging.ClusterLogging{},
	}

	if clusterLoggingRequest.createKibanaProxyClusterRoleBinding() != nil {
		t.Error("Exp. no error when creating proxy cluster role binding")
	}
}

func TestNewKibanaPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch")

	if len(podSpec.Containers) != 2 {
		t.Error("Exp. there to be 2 container")
	}

	// check pod node selector
	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contain the linux allocation selector but was %T", podSpec.NodeSelector)
	}
	if len(podSpec.NodeSelector) != 1 {
		t.Errorf("Exp. the nodeSelector to be %T but was %T", map[string]string{}, podSpec.NodeSelector)
	}
	if podSpec.NodeSelector[utils.OS_NODE_LABEL] != utils.LINUX_NODE_LABEL_VALUE {
		t.Errorf("Exp. the nodeSelector to contains %s: %s pair", utils.OS_NODE_LABEL, utils.LINUX_NODE_LABEL_VALUE)
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
		"foo":               "bar",
		utils.OS_NODE_LABEL: utils.LINUX_NODE_LABEL_VALUE,
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

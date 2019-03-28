package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"

	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestCreateKibanaProxyClusterRoleBindingWithoutError(t *testing.T) {
	var clusterRoleBinding *rbac.ClusterRoleBinding
	mockCreate := func(object sdk.Object) error {
		clusterRoleBinding = object.(*rbac.ClusterRoleBinding)
		return nil
	}
	cluster := NewClusterLogging(&logging.ClusterLogging{})
	cluster.Runtime.Create = mockCreate

	if cluster.createKibanaProxyClusterRoleBinding() != nil {
		t.Error("Exp. no error when creating proxy cluster role binding")
	}
	if clusterRoleBinding == nil {
		t.Error("Exp. clusterrolebinding not to be nil")
	}
	if len(clusterRoleBinding.GetOwnerReferences()) != 1 {
		t.Error("Exp. clusterrolebinding ownerreference to be set")
	}
	if clusterRoleBinding.RoleRef.Name != "system:auth-delegator" {
		t.Errorf("Exp. clusterrolebinding to be bound to system:auth-delegator but was %q", clusterRoleBinding.RoleRef.Name)
	}
	if len(clusterRoleBinding.Subjects) != 1 {
		t.Error("Exp. clusterrolebinding to be bound exactly one subject")
	}
	subject := clusterRoleBinding.Subjects[0]
	if subject.Name != "kibana" && subject.APIGroup != "" && subject.Namespace != "openshift-logging" {
		t.Errorf("Kibana proxy clusterrolebinding is not bound to the correct subject: %q", subject)
	}
}

func TestNewKibanaPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := NewClusterLogging(&logging.ClusterLogging{})
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
	if podSpec.NodeSelector != nil {
		t.Errorf("Exp. the nodeSelector to be %T but was %T", map[string]string{}, podSpec.NodeSelector)
	}
}

func TestNewKibanaPodSpecWhenResourcesAreDefined(t *testing.T) {
	cluster := NewClusterLogging(
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
	)
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
func TestNewKibanaPodSpecWhenNodeSelectorIsDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo": "bar",
	}
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				Visualization: logging.VisualizationSpec{
					Type: "kibana",
					KibanaSpec: logging.KibanaSpec{
						NodeSelector: expSelector,
					},
				},
			},
		},
	)
	podSpec := cluster.newKibanaPodSpec("test-app-name", "elasticsearch")

	//check kibana
	if !reflect.DeepEqual(podSpec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, podSpec.NodeSelector)
	}

}

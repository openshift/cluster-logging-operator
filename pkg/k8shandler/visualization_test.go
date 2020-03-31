package k8shandler

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewKibanaPodSpecSetsProxyToUseServiceAccountAsOAuthClient(t *testing.T) {
	clusterlogging := &logging.ClusterLogging{}
	spec := newKibanaPodSpec(clusterlogging, "kibana", "elasticsearch", nil, nil)
	for _, arg := range spec.Containers[1].Args {
		keyValue := strings.Split(arg, "=")
		if len(keyValue) >= 2 && keyValue[0] == "-client-id" {
			if keyValue[1] != "system:serviceaccount:openshift-logging:kibana" {
				t.Error("Exp. the proxy container arg 'client-id=system:serviceaccount:openshift-logging:kibana'")
			}
		}
		if len(keyValue) >= 2 && keyValue[0] == "-scope" {
			if keyValue[1] != "user:info user:check-access user:list-projects" {
				t.Error("Exp. the proxy container arg 'scope=user:info user:check-access user:list-projects'")
			}
		}
	}
}

func TestNewKibanaPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch", nil, nil)

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
			Visualization: &logging.VisualizationSpec{
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
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch", nil, nil)

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
			Visualization: &logging.VisualizationSpec{
				Type: "kibana",
				KibanaSpec: logging.KibanaSpec{
					NodeSelector: expSelector,
				},
			},
		},
	}
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "elasticsearch", nil, nil)

	//check kibana
	if !reflect.DeepEqual(podSpec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, podSpec.NodeSelector)
	}

}

func TestNewKibanaPodNoTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Visualization: &logging.VisualizationSpec{
				Type:       "kibana",
				KibanaSpec: logging.KibanaSpec{},
			},
		},
	}

	podSpec := newKibanaPodSpec(cluster, "test-app-name", "test-infra-name", nil, nil)
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
			Visualization: &logging.VisualizationSpec{
				Type: "kibana",
				KibanaSpec: logging.KibanaSpec{
					Tolerations: expTolerations,
				},
			},
		},
	}

	podSpec := newKibanaPodSpec(cluster, "test-app-name", "test-infra-name", nil, nil)
	tolerations := podSpec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewKibanaPodSpecWhenProxyConfigExists(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	httpproxy := "http://proxy-user@test.example.com/3128/"
	noproxy := ".cluster.local,localhost"
	caBundle := fmt.Sprint("-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n")
	podSpec := newKibanaPodSpec(cluster, "test-app-name", "test-infra-name",
		&configv1.Proxy{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Proxy",
				APIVersion: "config.openshift.io/v1",
			},
			Spec: configv1.ProxySpec{
				HTTPProxy:  httpproxy,
				HTTPSProxy: httpproxy,
				TrustedCA: configv1.ConfigMapNameReference{
					Name: "user-ca-bundle",
				},
			},
			Status: configv1.ProxyStatus{
				HTTPProxy:  httpproxy,
				HTTPSProxy: httpproxy,
				NoProxy:    noproxy,
			},
		},
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "openshift-logging",
				Name:      constants.KibanaTrustedCAName,
			},
			Data: map[string]string{
				constants.TrustedCABundleKey: caBundle,
			},
		},
	)

	if len(podSpec.Containers) != 2 {
		t.Error("Exp. there to be 2 kibana container")
	}

	checkKibanaProxyEnvVar(t, podSpec, "HTTP_PROXY", httpproxy)
	checkKibanaProxyEnvVar(t, podSpec, "HTTPS_PROXY", httpproxy)
	checkKibanaProxyEnvVar(t, podSpec, "NO_PROXY", noproxy)

	checkKibanaProxyVolumesAndVolumeMounts(t, podSpec, constants.KibanaTrustedCAName)
}

func checkKibanaProxyEnvVar(t *testing.T, podSpec v1.PodSpec, name string, value string) {
	env := podSpec.Containers[1].Env
	found := false
	for _, elem := range env {
		if elem.Name == name {
			found = true
			if elem.Value != value {
				t.Errorf("EnvVar %s: expected %s, actual %s", name, value, elem.Value)
			}
		}
	}
	if !found {
		t.Errorf("EnvVar %s not found", name)
	}
}

func checkKibanaProxyVolumesAndVolumeMounts(t *testing.T, podSpec v1.PodSpec, trustedca string) {
	volumemounts := podSpec.Containers[1].VolumeMounts
	found := false
	for _, elem := range volumemounts {
		if elem.Name == trustedca {
			found = true
			if elem.MountPath != constants.TrustedCABundleMountDir {
				t.Errorf("VolumeMounts %s: expected %s, actual %s", trustedca, constants.TrustedCABundleMountDir, elem.MountPath)
			}
		}
	}
	if !found {
		t.Errorf("VolumeMounts %s not found", trustedca)
	}

	volumes := podSpec.Volumes
	found = false
	for _, elem := range volumes {
		if elem.Name == trustedca {
			found = true
			if elem.VolumeSource.ConfigMap.LocalObjectReference.Name != trustedca {
				t.Errorf("Volume %s: expected %s, actual %s", trustedca, trustedca, elem.VolumeSource.Secret.SecretName)
			}
		}
	}
	if !found {
		t.Errorf("Volume %s not found", trustedca)
	}
}

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
		ForwardingRequest: nil,
		ForwardingSpec:    v1alpha1.ForwardingSpec{},
		Collector:         nil,
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

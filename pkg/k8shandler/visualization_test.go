package k8shandler

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
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

func TestNewLoggingSharedConfigMapExists(t *testing.T) {
	_ = routev1.AddToScheme(scheme.Scheme)
	cluster := &logging.ClusterLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: "openshift-logging",
		},
	}

	testCases := []struct {
		name    string
		objs    []runtime.Object
		wantCm  *v1.ConfigMap
		wantErr error
	}{
		{
			name: "new route creation",
			wantCm: NewConfigMap(
				loggingSharedConfigMapName,
				loggingSharedConfigNs,
				map[string]string{
					"kibanaAppPublicURL":      "https://",
					"kibanaInfraAppPublicURL": "https://",
				},
			),
		},
		{
			name: "update route with shared configmap, role and rolebinding migration",
			objs: []runtime.Object{
				runtime.Object(NewConfigMap(loggingSharedConfigMapNamePre44x, cluster.GetNamespace(), map[string]string{})),
				runtime.Object(NewRole(loggingSharedConfigRolePre44x, cluster.GetNamespace(), []rbac.PolicyRule{})),
				runtime.Object(NewRoleBinding(loggingSharedConfigRoleBindingPre44x, cluster.GetNamespace(), loggingSharedConfigRolePre44x, []rbac.Subject{})),
			},
			wantCm: NewConfigMap(
				loggingSharedConfigMapName,
				loggingSharedConfigNs,
				map[string]string{
					"kibanaAppPublicURL":      "https://",
					"kibanaInfraAppPublicURL": "https://",
				},
			),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewFakeClient(tc.objs...)
			clusterRequest := &ClusterLoggingRequest{
				client:  client,
				cluster: cluster,
			}

			if gotErr := clusterRequest.createOrUpdateKibanaRoute(); gotErr != tc.wantErr {
				t.Errorf("got: %v, want: %v", gotErr, tc.wantErr)
			}

			// Check new shared config map existings in openshift config shared namespace
			key := types.NamespacedName{Namespace: loggingSharedConfigNs, Name: loggingSharedConfigMapName}
			gotCm := &v1.ConfigMap{}
			utils.AddOwnerRefToObject(tc.wantCm, utils.AsOwner(clusterRequest.cluster))

			if err := client.Get(context.TODO(), key, gotCm); err != nil {
				t.Errorf("Expected configmap got: %v", err)
			}
			if ok := reflect.DeepEqual(gotCm, tc.wantCm); !ok {
				t.Errorf("got: %v, want: %v", gotCm, tc.wantCm)
			}

			// Check old shared config map is deleted
			key = types.NamespacedName{Namespace: cluster.GetNamespace(), Name: loggingSharedConfigMapNamePre44x}
			gotCmPre44x := &v1.ConfigMap{}
			if err := client.Get(context.TODO(), key, gotCmPre44x); !errors.IsNotFound(err) {
				t.Errorf("Expected deleted shared config pre 4.4.x, got: %v", err)
			}

			// Check old role to access the shared config map is deleted
			key = types.NamespacedName{Namespace: cluster.GetNamespace(), Name: loggingSharedConfigRolePre44x}
			gotRolePre44x := &rbac.Role{}
			if err := client.Get(context.TODO(), key, gotRolePre44x); !errors.IsNotFound(err) {
				t.Errorf("Expected deleted role for shared config map pre 4.4.x, got: %v", err)
			}

			// Check old rolebinding for group system:autheticated is deleted
			key = types.NamespacedName{Namespace: cluster.GetNamespace(), Name: loggingSharedConfigRoleBindingPre44x}
			gotRoleBindingPre44x := &rbac.RoleBinding{}
			if err := client.Get(context.TODO(), key, gotRoleBindingPre44x); !errors.IsNotFound(err) {
				t.Errorf("Expected deleted rolebinding for shared config map pre 4.4.x, got: %v", err)
			}
		})
	}
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

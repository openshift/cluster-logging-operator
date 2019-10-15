package k8shandler

import (
	"reflect"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewFluentdPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name", nil, logging.ForwardingSpec{})

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
			Collection: &logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
					},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name", nil, logging.ForwardingSpec{})

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
			Collection: &logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name", nil, logging.ForwardingSpec{})

	if !reflect.DeepEqual(podSpec.Tolerations, expectedTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expectedTolerations, podSpec.Tolerations)
	}
}

func TestNewFluentdPodSpecWhenSelectorIsDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo": "bar",
	}
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						NodeSelector: expSelector,
					},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name", nil, logging.ForwardingSpec{})

	if !reflect.DeepEqual(podSpec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, podSpec.NodeSelector)
	}
}

func TestNewFluentdPodNoTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{
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
			Collection: &logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type:        "fluentd",
					FluentdSpec: logging.FluentdSpec{},
				},
			},
		},
	}

	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name", nil, logging.ForwardingSpec{})
	tolerations := podSpec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewFluentdPodWithTolerations(t *testing.T) {

	providedToleration := v1.Toleration{
		Key:      "test",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}

	expTolerations := []v1.Toleration{
		providedToleration,
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
			Collection: &logging.CollectionSpec{
				logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						Tolerations: []v1.Toleration{
							providedToleration,
						},
					},
				},
			},
		},
	}

	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name", nil, logging.ForwardingSpec{})
	tolerations := podSpec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewFluentdPodSpecWhenProxyConfigExists(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	httpproxy := "http://proxy-user@test.example.com/3128/"
	noproxy := ".cluster.local,localhost"
	trustedca := "user-ca-bundle"
	podSpec := newFluentdPodSpec(cluster, "test-app-name", "test-infra-name",
		&configv1.Proxy{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Proxy",
				APIVersion: "config.openshift.io/v1",
			},
			Spec: configv1.ProxySpec{
				HTTPProxy:  httpproxy,
				HTTPSProxy: httpproxy,
				TrustedCA: configv1.ConfigMapNameReference{
					Name: trustedca,
				},
			},
			Status: configv1.ProxyStatus{
				HTTPProxy:  httpproxy,
				HTTPSProxy: httpproxy,
				NoProxy:    noproxy,
			},
		},
		logging.ForwardingSpec{},
	)

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}

	checkProxyEnvVar(t, podSpec, "HTTP_PROXY", httpproxy)
	checkProxyEnvVar(t, podSpec, "HTTPS_PROXY", httpproxy)
	checkProxyEnvVar(t, podSpec, "NO_PROXY", noproxy)

	checkProxyVolumesAndVolumeMounts(t, podSpec, trustedca)
}

func checkProxyEnvVar(t *testing.T, podSpec v1.PodSpec, name string, value string) {
	env := podSpec.Containers[0].Env
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

func checkProxyVolumesAndVolumeMounts(t *testing.T, podSpec v1.PodSpec, trustedca string) {
	name := "proxytrustedca"
	volumes := podSpec.Volumes
	found := false
	for _, elem := range volumes {
		if elem.Name == name {
			found = true
			if elem.VolumeSource.Secret.SecretName != trustedca {
				t.Errorf("Volume %s: expected %s, actual %s", name, trustedca, elem.VolumeSource.Secret.SecretName)
			}
		}
	}
	if !found {
		t.Errorf("Volume %s not found", name)
	}

	volumemounts := podSpec.Containers[0].VolumeMounts
	value := "/etc/fluent/proxy"
	found = false
	for _, elem := range volumemounts {
		if elem.Name == name {
			found = true
			if elem.MountPath != value {
				t.Errorf("VolumeMounts %s: expected %s, actual %s", name, value, elem.MountPath)
			}
		}
	}
	if !found {
		t.Errorf("VolumeMounts %s not found", name)
	}
}

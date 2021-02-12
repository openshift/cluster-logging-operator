package k8shandler

import (
	"os"
	"reflect"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("fluentd.go#newFluentdPodSpec", func() {
	var (
		cluster   = &logging.ClusterLogging{}
		podSpec   v1.PodSpec
		container v1.Container
	)
	BeforeEach(func() {
		podSpec = newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})
		container = podSpec.Containers[0]
	})
	Describe("when creating of the collector container", func() {

		It("should provide the pod IP as an environment var", func() {
			Expect(container.Env).To(IncludeEnvVar(v1.EnvVar{Name: "POD_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						APIVersion: "v1", FieldPath: "status.podIP"}}}))
		})
	})

})

func TestNewFluentdPodSpecWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})

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
	limitMemory := resource.MustParse("120Gi")
	requestMemory := resource.MustParse("100Gi")
	requestCPU := resource.MustParse("500m")
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						Resources: newResourceRequirements("120Gi", "", "100Gi", "500m"),
					},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})

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
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
		{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type: "fluentd",
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})

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
				Logs: logging.LogCollectionSpec{
					Type: "fluentd",
					FluentdSpec: logging.FluentdSpec{
						NodeSelector: expSelector,
					},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})

	if !reflect.DeepEqual(podSpec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, podSpec.NodeSelector)
	}
}

func TestNewFluentdPodNoTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
		{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type:        "fluentd",
					FluentdSpec: logging.FluentdSpec{},
				},
			},
		},
	}

	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})
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
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
		{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
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

	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})
	tolerations := podSpec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewFluentdPodSpecWhenProxyConfigExists(t *testing.T) {

	_httpProxy := os.Getenv("HTTP_PROXY")
	_httpsProxy := os.Getenv("HTTPS_PROXY")
	_noProxy := os.Getenv("NO_PROXY")
	cleanup := func() {
		_ = os.Setenv("HTTP_PROXY", _httpProxy)
		_ = os.Setenv("HTTPS_PROXY", _httpsProxy)
		_ = os.Setenv("NO_PROXY", _noProxy)
	}
	defer cleanup()

	cluster := &logging.ClusterLogging{}
	httpproxy := "http://proxy-user@test.example.com/3128/"
	noproxy := ".cluster.local,localhost"
	_ = os.Setenv("HTTP_PROXY", httpproxy)
	_ = os.Setenv("HTTPS_PROXY", httpproxy)
	_ = os.Setenv("NO_PROXY", noproxy)
	caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
	podSpec := newFluentdPodSpec(cluster, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "openshift-logging",
			Name:      constants.FluentdTrustedCAName,
		},
		Data: map[string]string{
			constants.TrustedCABundleKey: caBundle,
		},
	}, logging.ClusterLogForwarderSpec{})

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}

	checkFluentdProxyEnvVar(t, podSpec, "HTTP_PROXY", httpproxy)
	checkFluentdProxyEnvVar(t, podSpec, "HTTPS_PROXY", httpproxy)
	checkFluentdProxyEnvVar(t, podSpec, "NO_PROXY", noproxy)

	checkFluentdProxyVolumesAndVolumeMounts(t, podSpec, constants.FluentdTrustedCAName)
}

func TestFluentdPodInitContainerWithDefaultForwarding(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type:        "fluentd",
					FluentdSpec: logging.FluentdSpec{},
				},
			},
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 1,
				},
			},
		},
	}

	spec := logging.ClusterLogForwarderSpec{
		Outputs: []logging.OutputSpec{{Name: "default"}},
	}

	podSpec := newFluentdPodSpec(cluster, nil, spec)

	if len(podSpec.InitContainers) == 0 {
		t.Error("Expected pod to have defined init container")
	}
}

func TestFluentdPodInitContainerWithZeroESCount(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type:        "fluentd",
					FluentdSpec: logging.FluentdSpec{},
				},
			},
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
			},
		},
	}

	spec := logging.ClusterLogForwarderSpec{
		Outputs: []logging.OutputSpec{{Name: "default"}},
	}

	podSpec := newFluentdPodSpec(cluster, nil, spec)

	if len(podSpec.InitContainers) > 0 {
		t.Error("Expected pod not define init container")
	}
}

func TestFluentdPodNoInitContainerWithOutDefaultForwarding(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type:        "fluentd",
					FluentdSpec: logging.FluentdSpec{},
				},
			},
		},
	}

	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})
	if len(podSpec.InitContainers) > 0 {
		t.Error("Expected pod to have no init containers")
	}
}

func checkFluentdProxyEnvVar(t *testing.T, podSpec v1.PodSpec, name string, value string) {
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

func checkFluentdProxyVolumesAndVolumeMounts(t *testing.T, podSpec v1.PodSpec, trustedca string) {
	volumemounts := podSpec.Containers[0].VolumeMounts
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

func TestNewFluentdPodWhenChunkLimitSizeExists(t *testing.T) {
	chunkLimitSize, _ := utils.ParseQuantity("750k")
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Forwarder: &logging.ForwarderSpec{
				Fluentd: &logging.FluentdForwarderSpec{
					Buffer: &logging.FluentdBufferSpec{
						ChunkLimitSize: "750k",
					},
				},
			},
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type:        "fluentd",
					FluentdSpec: logging.FluentdSpec{},
				},
			},
		},
	}
	podSpec := newFluentdPodSpec(cluster, nil, logging.ClusterLogForwarderSpec{})

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 fluentd container")
	}

	checkFluentdForwarderEnvVar(t, podSpec, "BUFFER_SIZE_LIMIT", strconv.FormatInt(chunkLimitSize.Value(), 10))
}

func checkFluentdForwarderEnvVar(t *testing.T, podSpec v1.PodSpec, name string, value string) {
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

package collector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	"fmt"
	"os"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	vector "github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Factory#NewPodSpec", func() {
	var (
		podSpec   v1.PodSpec
		collector v1.Container

		factory *Factory
	)
	BeforeEach(func() {
		factory = &Factory{
			CollectorType: logging.LogCollectionTypeFluentd,
			Visit:         fluentd.CollectorVisitor,
		}
		utils.SetMockImageEnv() // NewPodSpec looks up image env vars.
		podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{})
		collector = podSpec.Containers[0]
	})
	Describe("when creating of the collector container", func() {

		It("should provide the pod IP as an environment var", func() {
			Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "POD_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						APIVersion: "v1", FieldPath: "status.podIP"}}}))
		})
		It("should set a security context", func() {
			Expect(collector.SecurityContext).To(Equal(&v1.SecurityContext{
				SELinuxOptions: &v1.SELinuxOptions{
					Type: "spc_t",
				},
				ReadOnlyRootFilesystem:   utils.GetBool(true),
				AllowPrivilegeEscalation: utils.GetBool(false),
			}))
		})
	})

	Describe("when creating the podSpec", func() {

		Context("and evaluating tolerations", func() {
			It("should add only defaults when none are defined", func() {
				Expect(podSpec.Tolerations).To(Equal(defaultTolerations))
			})

			It("should add the default and additional ones that are defined", func() {
				providedToleration := v1.Toleration{
					Key:      "test",
					Operator: v1.TolerationOpExists,
					Effect:   v1.TaintEffectNoSchedule,
				}
				factory.CollectorSpec = logging.CollectionSpec{
					Logs: logging.LogCollectionSpec{
						Type: "fluentd",
						FluentdSpec: logging.FluentdSpec{
							Tolerations: []v1.Toleration{
								providedToleration,
							},
						},
					},
				}
				podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{})
				expTolerations := append(defaultTolerations, providedToleration)
				Expect(podSpec.Tolerations).To(Equal(expTolerations))
			})

		})

		Context("and evaluating the node selector", func() {
			It("should add only defaults when none are defined", func() {
				Expect(podSpec.NodeSelector).To(Equal(utils.DefaultNodeSelector))
			})
			It("should add the selector when defined", func() {
				expSelector := map[string]string{
					"foo":             "bar",
					utils.OsNodeLabel: utils.LinuxValue,
				}
				factory.CollectorSpec = logging.CollectionSpec{
					Logs: logging.LogCollectionSpec{
						Type: "fluentd",
						FluentdSpec: logging.FluentdSpec{
							NodeSelector: map[string]string{
								"foo": "bar",
							},
						},
					},
				}
				podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{})
				Expect(podSpec.NodeSelector).To(Equal(expSelector))
			})

		})

		Context("and the proxy config exists", func() {

			var verifyEnvVar = func(container v1.Container, name, value string) {
				for _, elem := range container.Env {
					if elem.Name == name {
						Expect(elem.Value).To(Equal(value), "Exp. collector to have env var %s: %s:", name, value)
						return
					}
				}
				Fail(fmt.Sprintf("Exp. collector to include env var: %s", name))
			}

			var verifyProxyVolumesAndVolumeMounts = func(container v1.Container, podSpec v1.PodSpec, trustedca string) {
				found := false
				for _, elem := range container.VolumeMounts {
					if elem.Name == trustedca {
						found = true
						Expect(elem.MountPath).To(Equal(constants.TrustedCABundleMountDir), "VolumeMounts %s: expected %s, actual %s", trustedca, constants.TrustedCABundleMountDir, elem.MountPath)
						break
					}
				}
				if !found {
					Fail(fmt.Sprintf("Trusted ca-bundle VolumeMount %s not found for collector", trustedca))
				}

				for _, elem := range podSpec.Volumes {
					if elem.Name == trustedca {
						Expect(elem.VolumeSource.ConfigMap).To(Not(BeNil()), "Exp. the podSpec to have a mounted configmap for the trusted ca-bundle")
						Expect(elem.VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal(trustedca), "Volume %s: ConfigMap.LocalObjectReference.Name expected %s, actual %s", trustedca, elem.VolumeSource.ConfigMap.LocalObjectReference.Name, trustedca)
						return
					}
				}
				Fail(fmt.Sprintf("Volume %s not found for collector", trustedca))
			}

			It("should add the proxy variables to the collector", func() {
				_httpProxy := os.Getenv("HTTP_PROXY")
				_httpsProxy := os.Getenv("HTTPS_PROXY")
				_noProxy := os.Getenv("NO_PROXY")
				cleanup := func() {
					_ = os.Setenv("HTTP_PROXY", _httpProxy)
					_ = os.Setenv("HTTPS_PROXY", _httpsProxy)
					_ = os.Setenv("NO_PROXY", _noProxy)
				}
				defer cleanup()

				httpproxy := "http://proxy-user@test.example.com/3128/"
				noproxy := ".cluster.local,localhost"
				_ = os.Setenv("HTTP_PROXY", httpproxy)
				_ = os.Setenv("HTTPS_PROXY", httpproxy)
				_ = os.Setenv("NO_PROXY", noproxy)
				caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
				podSpec = *factory.NewPodSpec(&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "openshift-logging",
						Name:      constants.CollectorTrustedCAName,
					},
					Data: map[string]string{
						constants.TrustedCABundleKey: caBundle,
					},
				}, logging.ClusterLogForwarderSpec{})
				collector = podSpec.Containers[0]

				verifyEnvVar(collector, "HTTP_PROXY", httpproxy)
				verifyEnvVar(collector, "HTTPS_PROXY", httpproxy)
				verifyEnvVar(collector, "NO_PROXY", "elasticsearch,"+noproxy)
				verifyProxyVolumesAndVolumeMounts(collector, podSpec, constants.CollectorTrustedCAName)
			})
		})
	})

})

var _ = Describe("Factory#CollectorResourceRequirements", func() {
	var (
		factory *Factory
	)

	Context("when collectorType is vector", func() {
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeVector,
				Visit:         vector.CollectorVisitor,
			}
		})
		It("should not define any resources", func() {
			Expect(factory.CollectorResourceRequirements()).To(Equal(v1.ResourceRequirements{}))
		})

	})
	Context("when collectorType is fluentd", func() {
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeFluentd,
				Visit:         fluentd.CollectorVisitor,
			}
		})
		It("should apply the default resources when none are defined", func() {
			Expect(factory.CollectorResourceRequirements()).To(Equal(v1.ResourceRequirements{
				Limits: v1.ResourceList{v1.ResourceMemory: fluentd.DefaultMemory},
				Requests: v1.ResourceList{
					v1.ResourceMemory: fluentd.DefaultMemory,
					v1.ResourceCPU:    fluentd.DefaultCpuRequest,
				},
			}))
		})
		It("should apply the spec'd resources when defined", func() {
			factory.CollectorSpec = logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					FluentdSpec: logging.FluentdSpec{
						Resources: &v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("120Gi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Gi"),
								v1.ResourceCPU:    resource.MustParse("500m"),
							},
						},
					},
				},
			}
			Expect(factory.CollectorResourceRequirements()).To(Equal(v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("120Gi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("100Gi"),
					v1.ResourceCPU:    resource.MustParse("500m"),
				},
			}))
		})

	})
})

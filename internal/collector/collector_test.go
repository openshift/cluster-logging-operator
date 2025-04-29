package collector

import (
	"path"

	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	"fmt"
	"os"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	vector "github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	coreFactory "github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Factory#Daemonset", func() {
	Context("NewPodSpec", func() {
		var (
			podSpec   v1.PodSpec
			collector v1.Container

			factory *Factory
		)
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeFluentd,
				ImageName:     constants.FluentdName,
				Visit:         fluentd.CollectorVisitor,
				ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
				isDaemonset:   true,
				CommonLabelInitializer: func(o runtime.Object) {
					runtime.SetCommonLabels(o, string(logging.LogCollectionTypeFluentd), "test", constants.CollectorName)
				},
				PodLabelVisitor: func(o runtime.Object) {}, //do noting for fluentd
			}
			podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
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
					Capabilities: &v1.Capabilities{
						Drop: auth.RequiredDropCapabilities,
					},
					SELinuxOptions: &v1.SELinuxOptions{
						Type: "spc_t",
					},
					ReadOnlyRootFilesystem:   utils.GetPtr(true),
					AllowPrivilegeEscalation: utils.GetPtr(false),
					SeccompProfile: &v1.SeccompProfile{
						Type: v1.SeccompProfileTypeRuntimeDefault,
					},
				}))
			})
			Context("with custom ClusterLogForwarder name", func() {
				It("should have volumemount with custom name", func() {
					factory.ResourceNames = coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf"))
					expectedContainerVolume := v1.VolumeMount{
						Name:      "custom-clf-metrics",
						ReadOnly:  true,
						MountPath: metricsVolumePath}
					podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
					collector = podSpec.Containers[0]
					Expect(collector.VolumeMounts).To(ContainElement(expectedContainerVolume))
				})
			})
		})

		Describe("when creating the podSpec", func() {
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

			Context("and evaluating tolerations", func() {
				It("should add only defaults when none are defined", func() {
					Expect(podSpec.Tolerations).To(Equal(constants.DefaultTolerations()))
				})

				It("should add the default and additional ones that are defined", func() {
					providedToleration := v1.Toleration{
						Key:      "test",
						Operator: v1.TolerationOpExists,
						Effect:   v1.TaintEffectNoSchedule,
					}
					factory.CollectorSpec = logging.CollectionSpec{
						Type: "fluentd",
						CollectorSpec: logging.CollectorSpec{
							Tolerations: []v1.Toleration{
								providedToleration,
							},
						},
					}
					podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
					expTolerations := append(constants.DefaultTolerations(), providedToleration)
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
						Type: "fluentd",
						CollectorSpec: logging.CollectorSpec{
							NodeSelector: map[string]string{
								"foo": "bar",
							},
						},
					}
					podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
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

				DescribeTable("should add the proxy variables to the collector",
					func(collectorType logging.LogCollectionType, setHttpProxy, setHttpsProxy, setNoProxy, expectedHttpProxy, expectedHttpsProxy, expectedNoProxy string) {
						_httpProxy := os.Getenv(setHttpProxy)
						_httpsProxy := os.Getenv(setHttpsProxy)
						_noProxy := os.Getenv(setNoProxy)
						cleanup := func() {
							_ = os.Setenv(setHttpProxy, _httpProxy)
							_ = os.Setenv(setHttpsProxy, _httpsProxy)
							_ = os.Setenv(setNoProxy, _noProxy)
						}
						defer cleanup()

						httpproxy := "http://proxy-user@test.example.com/3128/"
						noproxy := ".cluster.local,localhost"
						_ = os.Setenv(setHttpProxy, httpproxy)
						_ = os.Setenv(setHttpsProxy, httpproxy)
						_ = os.Setenv(setNoProxy, noproxy)
						caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
						factory.CollectorSpec = logging.CollectionSpec{
							Type: collectorType,
						}
						podSpec = *factory.NewPodSpec(&v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "openshift-logging",
								Name:      constants.CollectorTrustedCAName,
							},
							Data: map[string]string{
								constants.TrustedCABundleKey: caBundle,
							},
						}, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
						collector = podSpec.Containers[0]

						verifyEnvVar(collector, expectedHttpProxy, httpproxy)
						verifyEnvVar(collector, expectedHttpsProxy, httpproxy)
						verifyEnvVar(collector, expectedNoProxy, "elasticsearch,"+noproxy)
						verifyProxyVolumesAndVolumeMounts(collector, podSpec, constants.CollectorTrustedCAName)
					},
					Entry("Fluentd expect environment variable name in lowercase", logging.LogCollectionTypeFluentd, "http_proxy", "https_proxy", "no_proxy", "http_proxy", "https_proxy", "no_proxy"),
					Entry("Fluentd expects existing uppercase environment variable values to be lowercased", logging.LogCollectionTypeFluentd, "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy"),
					Entry("Vector expect environment variable in uppercase", logging.LogCollectionTypeVector, "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"),
					Entry("Vector expects existing lowercase environment variable values to be uppercased", logging.LogCollectionTypeVector, "http_proxy", "https_proxy", "no_proxy", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"),
				)
			})

			Context("and using custom named ClusterLogForwarder", func() {

				It("should have custom named podSpec resources based on CLF name", func() {
					clf := *runtime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf")
					clf.Spec.ServiceAccountName = "custom-clf"
					factory.ResourceNames = coreFactory.GenerateResourceNames(clf)
					expectedPodSpecMetricsVol := v1.Volume{
						Name: "custom-clf-metrics",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "custom-clf-metrics",
							}}}

					caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
					podSpec = *factory.NewPodSpec(&v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "openshift-logging",
							Name:      factory.ResourceNames.CaTrustBundle,
						},
						Data: map[string]string{
							constants.TrustedCABundleKey: caBundle,
						},
					}, logging.ClusterLogForwarderSpec{}, "foobar", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)

					collector = podSpec.Containers[0]
					Expect(podSpec.Volumes).To(ContainElement(expectedPodSpecMetricsVol))
					Expect(podSpec.ServiceAccountName).To(Equal("custom-clf"))
					verifyProxyVolumesAndVolumeMounts(collector, podSpec, "custom-clf-trustbundle")
				})
			})

			Context("and mounting volumes", func() {
				It("should mount host path volumes", func() {
					Expect(podSpec.Volumes).To(HaveLen(15))
					Expect(podSpec.Volumes).To(ContainElement(v1.Volume{Name: logContainers, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logContainersValue}}}))
				})
			})
		})
		Context("NewDaemonset", func() {
			const targetAnnotation = "target.workload.openshift.io/management"
			It("should have required annotations", func() {
				actDs := factory.NewDaemonSet(constants.OpenshiftNS, "test", nil, tls.GetClusterTLSProfileSpec(nil), nil)
				Expect(actDs.Spec.Template.Annotations).To(HaveKey(constants.AnnotationSecretHash))
				Expect(actDs.Spec.Template.Annotations).To(HaveKey(targetAnnotation))
			})
		})
	})
	Context("CollectorContainerVolumeMounts", func() {

		DescribeTable("when CLF specs inputs", func(inputs []logging.InputSpec, numVolumeMounts int, expVolumeMounts []v1.VolumeMount) {
			factory := &Factory{
				CollectorType: logging.LogCollectionTypeFluentd,
				ImageName:     constants.FluentdName,
				Visit:         fluentd.CollectorVisitor,
				ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
				isDaemonset:   true,
				CommonLabelInitializer: func(o runtime.Object) {
					runtime.SetCommonLabels(o, string(logging.LogCollectionTypeFluentd), "test", constants.CollectorName)
				},
				PodLabelVisitor: func(o runtime.Object) {},
			}
			podSpec := *factory.NewPodSpec(nil,
				logging.ClusterLogForwarderSpec{
					Inputs: inputs,
				},
				"1234",
				"",
				tls.GetClusterTLSProfileSpec(nil),
				nil,
				constants.OpenshiftNS)
			collector := podSpec.Containers[0]

			// Check VolumeMounts
			Expect(collector.VolumeMounts).To(HaveLen(numVolumeMounts))
			for _, vm := range expVolumeMounts {
				Expect(collector.VolumeMounts).To(ContainElement(vm))
			}
		},
			Entry("should mount all volumes if all reserved input types are spec'd empty",
				[]logging.InputSpec{
					{Application: &logging.Application{}},
					{Infrastructure: &logging.Infrastructure{}},
					{Audit: &logging.Audit{}},
				},
				15,
				[]v1.VolumeMount{
					{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
					{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
					{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
					{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
					{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
					{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
					{Name: logOauthserver, ReadOnly: true, MountPath: logOauthserverValue},
					{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
					{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
				}),
			Entry("should only mount container/pod sources when application is spec'd empty",
				[]logging.InputSpec{
					{Application: &logging.Application{}},
				},
				8,
				[]v1.VolumeMount{
					{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
					{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
				}),
			Entry("should mount journal & container sources when infrastructure is spec'd empty",
				[]logging.InputSpec{
					{Infrastructure: &logging.Infrastructure{}},
				},
				9,
				[]v1.VolumeMount{
					{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
					{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
					{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
				}),
			Entry("should only mount container sources when infrastructure is spec'd with container source",
				[]logging.InputSpec{
					{Infrastructure: &logging.Infrastructure{Sources: []string{logging.InfrastructureSourceContainer}}},
				},
				8,
				[]v1.VolumeMount{
					{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
					{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
				}),
			Entry("should only mount journal sources when infrastructure is spec'd with node source",
				[]logging.InputSpec{
					{Infrastructure: &logging.Infrastructure{Sources: []string{logging.InfrastructureSourceNode}}},
				},
				7,
				[]v1.VolumeMount{
					{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
				}),
			Entry("should mount all audit sources when audit is spec'd empty",
				[]logging.InputSpec{
					{Audit: &logging.Audit{}},
				},
				12,
				[]v1.VolumeMount{
					{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
					{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
					{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
					{Name: logOauthserver, ReadOnly: true, MountPath: logOauthserverValue},
					{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
					{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
				}),
			Entry("should mount audit logs when audit is spec'd with auditd source",
				[]logging.InputSpec{
					{Audit: &logging.Audit{Sources: []string{logging.AuditSourceAuditd}}},
				},
				7,
				[]v1.VolumeMount{
					{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
				}),
			Entry("should mount kubeApi logs when audit is spec'd with kubeAPI source",
				[]logging.InputSpec{
					{Audit: &logging.Audit{Sources: []string{logging.AuditSourceKube}}},
				},
				7,
				[]v1.VolumeMount{
					{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
				}),
			Entry("should mount openshiftAPI logs when audit is spec'd with openshiftAPI source",
				[]logging.InputSpec{
					{Audit: &logging.Audit{Sources: []string{logging.AuditSourceOpenShift}}},
				},
				9,
				[]v1.VolumeMount{
					{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
					{Name: logOauthserver, ReadOnly: true, MountPath: logOauthserverValue},
					{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
				}),
			Entry("should mount OVN logs when audit is spec'd with OVN source",
				[]logging.InputSpec{
					{Audit: &logging.Audit{Sources: []string{logging.AuditSourceOVN}}},
				},
				7,
				[]v1.VolumeMount{
					{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
				}),
			Entry("should mount journal and openshiftAPI sources if infra is spec'd with node and audit spec'd with openshiftAPI",
				[]logging.InputSpec{
					{Infrastructure: &logging.Infrastructure{Sources: []string{logging.InfrastructureSourceNode}}},
					{Audit: &logging.Audit{Sources: []string{logging.AuditSourceOpenShift}}},
				},
				10,
				[]v1.VolumeMount{
					{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
					{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
					{Name: logOauthserver, ReadOnly: true, MountPath: logOauthserverValue},
					{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
				}),
			Entry("should mount container and kubeAPI sources if application is spec'd and audit spec'd with kubeAPI",
				[]logging.InputSpec{
					{Application: &logging.Application{}},
					{Audit: &logging.Audit{Sources: []string{logging.AuditSourceKube}}},
				},
				9,
				[]v1.VolumeMount{
					{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
					{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
					{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
				}))
	})
})

var _ = Describe("Factory#Deployment", func() {
	var (
		podSpec   v1.PodSpec
		collector v1.Container

		factory *Factory
	)
	BeforeEach(func() {
		factory = &Factory{
			CollectorType: logging.LogCollectionTypeFluentd,
			ImageName:     constants.FluentdName,
			Visit:         fluentd.CollectorVisitor,
			ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
			isDaemonset:   false,
			CommonLabelInitializer: func(o runtime.Object) {
				runtime.SetCommonLabels(o, string(logging.LogCollectionTypeFluentd), "test", constants.CollectorName)
			},
			PodLabelVisitor: func(o runtime.Object) {}, //do noting for fluentd
		}
		podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
		collector = podSpec.Containers[0]
	})
	Context("NewPodSpec", func() {
		Describe("when creating of the collector container", func() {

			It("should provide the pod IP as an environment var", func() {
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "POD_IP",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{
							APIVersion: "v1", FieldPath: "status.podIP"}}}))
			})
			It("should not set security context", func() {
				Expect(collector.SecurityContext).ToNot(Equal(&v1.SecurityContext{
					Capabilities: &v1.Capabilities{
						Drop: auth.RequiredDropCapabilities,
					},
					SELinuxOptions: &v1.SELinuxOptions{
						Type: "spc_t",
					},
					ReadOnlyRootFilesystem:   utils.GetPtr(true),
					AllowPrivilegeEscalation: utils.GetPtr(false),
					SeccompProfile: &v1.SeccompProfile{
						Type: v1.SeccompProfileTypeRuntimeDefault,
					},
				}))
			})
			Context("with custom ClusterLogForwarder name", func() {
				It("should have volumemount with custom name", func() {
					factory.ResourceNames = coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf"))
					expectedContainerVolume := v1.VolumeMount{
						Name:      "custom-clf-metrics",
						ReadOnly:  true,
						MountPath: metricsVolumePath}
					podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
					collector = podSpec.Containers[0]
					Expect(collector.VolumeMounts).To(ContainElement(expectedContainerVolume))
				})
			})
		})

		Describe("when creating the podSpec", func() {
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

			Context("and evaluating tolerations", func() {
				It("should add only defaults when none are defined", func() {
					Expect(podSpec.Tolerations).To(Equal(constants.DefaultTolerations()))
				})

				It("should add the default and additional ones that are defined", func() {
					providedToleration := v1.Toleration{
						Key:      "test",
						Operator: v1.TolerationOpExists,
						Effect:   v1.TaintEffectNoSchedule,
					}
					factory.CollectorSpec = logging.CollectionSpec{
						Type: "fluentd",
						CollectorSpec: logging.CollectorSpec{
							Tolerations: []v1.Toleration{
								providedToleration,
							},
						},
					}
					podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
					expTolerations := append(constants.DefaultTolerations(), providedToleration)
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
						Type: "fluentd",
						CollectorSpec: logging.CollectorSpec{
							NodeSelector: map[string]string{
								"foo": "bar",
							},
						},
					}
					podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
					Expect(podSpec.NodeSelector).To(Equal(expSelector))
				})

			})

			Context("and using custom named ClusterLogForwarder", func() {

				It("should have custom named podSpec resources based on CLF name", func() {
					clf := *runtime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf")
					clf.Spec.ServiceAccountName = "custom-clf"
					factory.ResourceNames = coreFactory.GenerateResourceNames(clf)
					expectedPodSpecMetricsVol := v1.Volume{
						Name: "custom-clf-metrics",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "custom-clf-metrics",
							}}}

					caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
					podSpec = *factory.NewPodSpec(&v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "openshift-logging",
							Name:      factory.ResourceNames.CaTrustBundle,
						},
						Data: map[string]string{
							constants.TrustedCABundleKey: caBundle,
						},
					}, logging.ClusterLogForwarderSpec{}, "foobar", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)

					collector = podSpec.Containers[0]
					Expect(podSpec.Volumes).To(ContainElement(expectedPodSpecMetricsVol))
					Expect(podSpec.ServiceAccountName).To(Equal("custom-clf"))
					verifyProxyVolumesAndVolumeMounts(collector, podSpec, "custom-clf-trustbundle")
				})
			})

			Context("and mounting volumes", func() {
				It("should not mount host path volumes", func() {
					Expect(podSpec.Volumes).To(HaveLen(6))
					Expect(podSpec.Volumes).NotTo(ContainElement(v1.Volume{Name: logContainers, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logContainersValue}}}))
				})
			})
		})
	})

	Context("NewDeployment", func() {
		const targetAnnotation = "target.workload.openshift.io/management"
		It("should have required annotations", func() {
			actDpl := factory.NewDeployment(constants.OpenshiftNS, "test", nil, tls.GetClusterTLSProfileSpec(nil), nil)
			Expect(actDpl.Spec.Template.Annotations).To(HaveKey(constants.AnnotationSecretHash))
			Expect(actDpl.Spec.Template.Annotations).To(HaveKey(targetAnnotation))
		})
	})

})

var _ = Describe("Factory#CollectorResourceRequirements", func() {
	var (
		factory        *Factory
		collectionSpec = logging.CollectionSpec{
			CollectorSpec: logging.CollectorSpec{
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
		}
		expResources = v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("120Gi"),
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("100Gi"),
				v1.ResourceCPU:    resource.MustParse("500m"),
			},
		}
	)

	Context("when collectorType is vector", func() {
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeVector,
				ImageName:     constants.VectorName,
				Visit:         vector.CollectorVisitor,
			}
		})
		It("should not define any resources when none are specified", func() {
			Expect(factory.CollectorResourceRequirements()).To(Equal(v1.ResourceRequirements{}))
		})

		It("should apply the spec'd resources when defined", func() {
			factory.CollectorSpec = collectionSpec
			Expect(factory.CollectorResourceRequirements()).To(Equal(expResources))
		})

	})
	Context("when collectorType is fluentd", func() {
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeFluentd,
				ImageName:     constants.FluentdName,
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
			factory.CollectorSpec = collectionSpec
			Expect(factory.CollectorResourceRequirements()).To(Equal(expResources))
		})

	})
})

var _ = Describe("Factory#NewPodSpec Add Cloudwatch STS Resources", func() {
	var (
		factory   *Factory
		pipelines = []logging.PipelineSpec{
			{
				Name:       "cw-forward",
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"cw"},
			},
		}
		outputs = []logging.OutputSpec{
			{
				Type: logging.OutputTypeCloudwatch,
				Name: "cw",
				OutputTypeSpec: logging.OutputTypeSpec{
					Cloudwatch: &logging.Cloudwatch{
						Region:  "us-east-77",
						GroupBy: logging.LogGroupByNamespaceName,
					},
				},
				Secret: &logging.OutputSecretSpec{
					Name: "my-secret",
				},
			},
		}
		roleArn = "arn:aws:iam::123456789012:role/my-role-to-assume"
		secrets = map[string]*v1.Secret{
			// output secrets are keyed by output name
			outputs[0].Name: {
				Data: map[string][]byte{
					"credentials": []byte(roleArn),
				},
			},
		}
	)
	Context("when collectorType is fluentd", func() {
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeFluentd,
				ImageName:     constants.FluentdName,
				Visit:         fluentd.CollectorVisitor,
				Secrets:       secrets,
				ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
			}
		})
		Context("when collector has a secret containing a credentials key", func() {

			It("should NO LONGER be setting AWS ENV vars in the container", func() {
				podSpec := *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{
					Outputs:   outputs,
					Pipelines: pipelines,
				}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector := podSpec.Containers[0]

				// LOG-4084 fluentd no longer setting env vars
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSRegionEnvVarKey,
				})))
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSRoleArnEnvVarKey,
				})))
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSRoleSessionEnvVarKey,
				})))
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSWebIdentityTokenEnvVarKey,
				})))
			})
		})
		Context("when collector has a secret containing a role_arn key", func() {
			BeforeEach(func() {
				factory.Secrets = map[string]*v1.Secret{
					outputs[0].Name: {
						Data: map[string][]byte{
							"role_arn": []byte(roleArn),
						},
					},
				}
			})
			It("should NO LONGER be setting AWS ENV vars in the container", func() {
				podSpec := *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{
					Outputs:   outputs,
					Pipelines: pipelines,
				}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector := podSpec.Containers[0]

				// LOG-4084 fluentd no longer setting env vars
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSRegionEnvVarKey,
				})))
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSRoleArnEnvVarKey,
				})))
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSRoleSessionEnvVarKey,
				})))
				Expect(collector.Env).To(Not(IncludeEnvVar(v1.EnvVar{
					Name: constants.AWSWebIdentityTokenEnvVarKey,
				})))
			})
		})
	})
	Context("when collectorType is vector", func() {
		BeforeEach(func() {
			factory = &Factory{
				CollectorType: logging.LogCollectionTypeVector,
				ImageName:     constants.VectorName,
				Visit:         vector.CollectorVisitor,
				Secrets:       secrets,
				ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
			}
		})
		Context("when collector has a secret containing a credentials key", func() {

			It("should find the AWS web identity env vars in the container", func() {
				podSpec := *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{
					Outputs:   outputs,
					Pipelines: pipelines,
				}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector := podSpec.Containers[0]

				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSRegionEnvVarKey,
					Value: outputs[0].OutputTypeSpec.Cloudwatch.Region,
				}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSRoleArnEnvVarKey,
					Value: roleArn,
				}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSRoleSessionEnvVarKey,
					Value: constants.AWSRoleSessionName,
				}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSWebIdentityTokenEnvVarKey,
					Value: path.Join(constants.AWSWebIdentityTokenMount, constants.AWSWebIdentityTokenFilePath),
				}))
			})
		})
		Context("when collector has a secret containing a role_arn key", func() {
			BeforeEach(func() {
				factory.Secrets = map[string]*v1.Secret{
					outputs[0].Name: {
						Data: map[string][]byte{
							"role_arn": []byte(roleArn),
						},
					},
				}
			})
			It("should find the AWS web identity env vars in the container", func() {
				podSpec := *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{
					Outputs:   outputs,
					Pipelines: pipelines,
				}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector := podSpec.Containers[0]

				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSRegionEnvVarKey,
					Value: outputs[0].OutputTypeSpec.Cloudwatch.Region,
				}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSRoleArnEnvVarKey,
					Value: roleArn,
				}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSRoleSessionEnvVarKey,
					Value: constants.AWSRoleSessionName,
				}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
					Name:  constants.AWSWebIdentityTokenEnvVarKey,
					Value: path.Join(constants.AWSWebIdentityTokenMount, constants.AWSWebIdentityTokenFilePath),
				}))

			})
		})
	})
})

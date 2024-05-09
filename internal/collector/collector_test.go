package collector

import (
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	"fmt"
	"os"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	vector "github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	coreFactory "github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Factory#Daemonset#NewPodSpec", func() {
	var (
		podSpec   v1.PodSpec
		collector v1.Container

		factory *Factory
	)
	BeforeEach(func() {
		factory = &Factory{
			CollectorType: logging.LogCollectionTypeVector,
			ImageName:     constants.VectorName,
			Visit:         vector.CollectorVisitor,
			ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
			isDaemonset:   true,
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

		Context("with vector collector", func() {
			It("should set VECTOR_LOG env variable with debug value", func() {
				logLevelDebug := "debug"
				factory.CollectorType = logging.LogCollectionTypeVector
				factory.ImageName = constants.VectorName
				factory.Visit = vector.CollectorVisitor
				factory.LogLevel = logLevelDebug

				podSpec = *factory.NewPodSpec(nil, logging.ClusterLogForwarderSpec{}, "1234", "", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector = podSpec.Containers[0]
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "VECTOR_LOG", Value: logLevelDebug}))
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

			It("should add the proxy variables to the collector", func() {
				_httpProxy := os.Getenv("http_proxy")
				_httpsProxy := os.Getenv("https_proxy")
				_noProxy := os.Getenv("no_proxy")
				cleanup := func() {
					_ = os.Setenv("http_proxy", _httpProxy)
					_ = os.Setenv("https_proxy", _httpsProxy)
					_ = os.Setenv("no_proxy", _noProxy)
				}
				defer cleanup()

				httpproxy := "http://proxy-user@test.example.com/3128/"
				noproxy := ".cluster.local,localhost"
				_ = os.Setenv("http_proxy", httpproxy)
				_ = os.Setenv("https_proxy", httpproxy)
				_ = os.Setenv("no_proxy", noproxy)
				caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
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

				verifyEnvVar(collector, "http_proxy", httpproxy)
				verifyEnvVar(collector, "https_proxy", httpproxy)
				verifyEnvVar(collector, "no_proxy", "elasticsearch,"+noproxy)
				verifyProxyVolumesAndVolumeMounts(collector, podSpec, constants.CollectorTrustedCAName)
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
			It("should mount host path volumes", func() {
				Expect(podSpec.Volumes).To(ContainElement(v1.Volume{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}}))
				Expect(podSpec.Volumes).To(HaveLen(13))
			})
		})
	})

})

var _ = Describe("Factory#Deployment#NewPodSpec", func() {
	var (
		podSpec   v1.PodSpec
		collector v1.Container

		factory *Factory
	)
	BeforeEach(func() {
		factory = &Factory{
			CollectorType: logging.LogCollectionTypeVector,
			ImageName:     constants.VectorName,
			Visit:         vector.CollectorVisitor,
			ResourceNames: coreFactory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
			isDaemonset:   false,
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

			It("should add the proxy variables to the collector", func() {
				_httpProxy := os.Getenv("http_proxy")
				_httpsProxy := os.Getenv("https_proxy")
				_noProxy := os.Getenv("no_proxy")
				cleanup := func() {
					_ = os.Setenv("http_proxy", _httpProxy)
					_ = os.Setenv("https_proxy", _httpsProxy)
					_ = os.Setenv("no_proxy", _noProxy)
				}
				defer cleanup()

				httpproxy := "http://proxy-user@test.example.com/3128/"
				noproxy := ".cluster.local,localhost"
				_ = os.Setenv("http_proxy", httpproxy)
				_ = os.Setenv("https_proxy", httpproxy)
				_ = os.Setenv("no_proxy", noproxy)
				caBundle := "-----BEGIN CERTIFICATE-----\n<PEM_ENCODED_CERT>\n-----END CERTIFICATE-----\n"
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

				verifyEnvVar(collector, "http_proxy", httpproxy)
				verifyEnvVar(collector, "https_proxy", httpproxy)
				verifyEnvVar(collector, "no_proxy", "elasticsearch,"+noproxy)
				verifyProxyVolumesAndVolumeMounts(collector, podSpec, constants.CollectorTrustedCAName)
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
				Expect(podSpec.Volumes).NotTo(ContainElement(v1.Volume{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}}))
				Expect(podSpec.Volumes).To(HaveLen(5))
			})
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
				//TODO: FIXME
				//Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				//	Name:  constants.AWSRoleArnEnvVarKey,
				//	Value: roleArn,
				//}))
				//Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				//	Name:  constants.AWSRoleSessionEnvVarKey,
				//	Value: constants.AWSRoleSessionName,
				//}))
				//Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				//	Name:  constants.AWSWebIdentityTokenEnvVarKey,
				//	Value: path.Join(constants.AWSWebIdentityTokenMount, constants.AWSWebIdentityTokenFilePath),
				//}))
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
				//TODO: FIXME
				//Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				//	Name:  constants.AWSRoleArnEnvVarKey,
				//	Value: roleArn,
				//}))
				//Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				//	Name:  constants.AWSRoleSessionEnvVarKey,
				//	Value: constants.AWSRoleSessionName,
				//}))
				//Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				//	Name:  constants.AWSWebIdentityTokenEnvVarKey,
				//	Value: path.Join(constants.AWSWebIdentityTokenMount, constants.AWSWebIdentityTokenFilePath),
				//}))

			})
		})
	})
})

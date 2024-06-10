package collector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	vector "github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	coreFactory "github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"os"

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
			ImageName:     constants.VectorName,
			Visit:         vector.CollectorVisitor,
			ResourceNames: coreFactory.ResourceNames(*obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName, runtime.Initialize)),
			isDaemonset:   true,
		}
		podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{
			Outputs: []obs.OutputSpec{
				{
					Name: "myloki",
					Type: obs.OutputTypeLokiStack,
					TLS: &obs.OutputTLSSpec{
						TLSSpec: obs.TLSSpec{
							CA: &obs.ConfigMapOrSecretKey{
								Key: "myca",
								ConfigMap: &v1.LocalObjectReference{
									Name: "bar",
								},
							},
						},
					},
					LokiStack: &obs.LokiStack{
						Authentication: &obs.HTTPAuthentication{
							Token: &obs.BearerToken{
								From: obs.BearerTokenFromServiceAccountToken,
							},
							Username: &obs.SecretKey{
								Key: "foo",
								Secret: &v1.LocalObjectReference{
									Name: "bar",
								},
							},
						},
					},
				},
			},
		}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
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
		Context("the volume for metrics certificates", func() {
			It("should mount a volume  based upon the forwarder name", func() {
				factory.ResourceNames = coreFactory.ResourceNames(*obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf", runtime.Initialize))
				expectedContainerVolume := v1.VolumeMount{
					Name:      "custom-clf-metrics",
					ReadOnly:  true,
					MountPath: metricsVolumePath}
				podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector = podSpec.Containers[0]
				Expect(collector.VolumeMounts).To(ContainElement(expectedContainerVolume))
			})
		})

		It("should set VECTOR_LOG env variable with debug value", func() {
			logLevelDebug := "debug"
			factory.LogLevel = logLevelDebug

			podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
			collector = podSpec.Containers[0]
			Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "VECTOR_LOG", Value: logLevelDebug}))
		})

		Context("the volume mounts", func() {
			It("should mount all output configmaps", func() {
				Expect(collector.VolumeMounts).To(IncludeVolumeMount(
					v1.VolumeMount{
						Name:      "config-bar",
						ReadOnly:  true,
						MountPath: common.ConfigMapBasePath("bar")}))
			})
			It("should mount all output secrets", func() {
				Expect(collector.VolumeMounts).To(IncludeVolumeMount(
					v1.VolumeMount{
						Name:      "bar",
						ReadOnly:  true,
						MountPath: common.SecretBasePath("bar")}))
			})
			It("should mount the service account projected token", func() {
				Expect(collector.VolumeMounts).To(IncludeVolumeMount(
					v1.VolumeMount{
						Name:      saTokenVolumeName,
						ReadOnly:  true,
						MountPath: common.ServiceAccountBasePath(saTokenVolumeName)}))
			})
		})
	})

	Describe("when creating the podSpec", func() {

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
				factory.CollectorSpec = obs.CollectorSpec{
					Tolerations: []v1.Toleration{
						providedToleration,
					},
				}
				podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
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
				factory.CollectorSpec = obs.CollectorSpec{
					NodeSelector: map[string]string{
						"foo": "bar",
					},
				}
				podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				Expect(podSpec.NodeSelector).To(Equal(expSelector))
			})

		})

		Context("and the proxy config exists", func() {

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
				}, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector = podSpec.Containers[0]

				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "http_proxy", Value: httpproxy}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "https_proxy", Value: httpproxy}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "no_proxy", Value: "elasticsearch," + noproxy}))
				Expect(collector.VolumeMounts).To(IncludeVolumeMount(v1.VolumeMount{Name: constants.VolumeNameTrustedCA,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}))
				Expect(podSpec.Volumes).To(IncludeVolume(v1.Volume{
					Name: constants.VolumeNameTrustedCA,
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: constants.CollectorTrustedCAName,
							},
							Items: []v1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}))
			})
		})

		It("should have podSpec attribute names based on CLF name", func() {
			clf := *obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf", runtime.Initialize)
			clf.Spec.ServiceAccount.Name = "custom-clf"
			factory.ResourceNames = coreFactory.ResourceNames(clf)
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
					Name:      constants.CollectorTrustedCAName,
				},
				Data: map[string]string{
					constants.TrustedCABundleKey: caBundle,
				},
			}, obs.ClusterLogForwarderSpec{}, "foobar", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)

			collector = podSpec.Containers[0]
			Expect(podSpec.Volumes).To(ContainElement(expectedPodSpecMetricsVol))
			Expect(podSpec.ServiceAccountName).To(Equal("custom-clf"))
			Expect(collector.VolumeMounts).To(IncludeVolumeMount(v1.VolumeMount{Name: constants.VolumeNameTrustedCA,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			}))
			Expect(podSpec.Volumes).To(IncludeVolume(v1.Volume{
				Name: constants.VolumeNameTrustedCA,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: constants.CollectorTrustedCAName,
						},
						Items: []v1.KeyToPath{
							{
								Key:  constants.TrustedCABundleKey,
								Path: constants.TrustedCABundleMountFile,
							},
						},
					},
				},
			}))
		})

		Context("and mounting volumes", func() {
			It("should mount host path volumes", func() {
				Expect(podSpec.Volumes).To(IncludeVolume(v1.Volume{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}}))
				Expect(podSpec.Volumes).To(HaveLen(16))
			})

			It("should mount all volumes for output configmaps", func() {
				Expect(podSpec.Volumes).To(IncludeVolume(
					v1.Volume{
						Name: "config-bar",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
							},
						},
					}))
			})
			It("should mount all volumes for output secrets", func() {
				Expect(podSpec.Volumes).To(IncludeVolume(
					v1.Volume{
						Name: "bar",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "bar",
							},
						},
					}))
			})
			It("should mount the service account projected token", func() {
				Expect(podSpec.Volumes).To(IncludeVolume(
					v1.Volume{
						Name: saTokenVolumeName,
						VolumeSource: v1.VolumeSource{
							Projected: &v1.ProjectedVolumeSource{
								Sources: []v1.VolumeProjection{
									{
										ServiceAccountToken: &v1.ServiceAccountTokenProjection{
											Audience:          "openshift",
											ExpirationSeconds: utils.GetPtr[int64](saTokenExpirationSecs),
											Path:              constants.TokenKey,
										},
									},
								},
							},
						},
					}))
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
			ImageName:     constants.VectorName,
			Visit:         vector.CollectorVisitor,
			ResourceNames: coreFactory.ResourceNames(*obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName, runtime.Initialize)),
			isDaemonset:   false,
		}
		podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
		collector = podSpec.Containers[0]
	})
	Describe("when creating the collector container", func() {

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
		It("should mount a volume for metrics certificates based upon the CR name", func() {
			factory.ResourceNames = coreFactory.ResourceNames(*obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf", runtime.Initialize))
			expectedContainerVolume := v1.VolumeMount{
				Name:      "custom-clf-metrics",
				ReadOnly:  true,
				MountPath: metricsVolumePath}
			podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
			collector = podSpec.Containers[0]
			Expect(collector.VolumeMounts).To(ContainElement(expectedContainerVolume))
		})
	})

	Describe("when creating the podSpec", func() {

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
				factory.CollectorSpec = obs.CollectorSpec{
					Tolerations: []v1.Toleration{
						providedToleration,
					},
				}
				podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
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
				factory.CollectorSpec = obs.CollectorSpec{
					NodeSelector: map[string]string{
						"foo": "bar",
					},
				}
				podSpec = *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				Expect(podSpec.NodeSelector).To(Equal(expSelector))
			})

		})

		Context("and the proxy config exists", func() {

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
						Name:      factory.ResourceNames.CaTrustBundle,
					},
					Data: map[string]string{
						constants.TrustedCABundleKey: caBundle,
					},
				}, obs.ClusterLogForwarderSpec{}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
				collector = podSpec.Containers[0]

				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "http_proxy", Value: httpproxy}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "https_proxy", Value: httpproxy}))
				Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{Name: "no_proxy", Value: "elasticsearch," + noproxy}))
				Expect(podSpec.Volumes).To(IncludeVolume(v1.Volume{
					Name: constants.VolumeNameTrustedCA,
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: factory.ResourceNames.CaTrustBundle,
							},
							Items: []v1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}))
				Expect(podSpec.Containers[0].VolumeMounts).To(IncludeVolumeMount(v1.VolumeMount{
					Name:      constants.VolumeNameTrustedCA,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}))
			})
		})

		Context("and using custom named ClusterLogForwarder", func() {

			It("should have custom named podSpec resources based on CLF name", func() {
				clf := *obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, "custom-clf", runtime.Initialize)
				clf.Spec.ServiceAccount.Name = "custom-clf"
				factory.ResourceNames = coreFactory.ResourceNames(clf)
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
				}, obs.ClusterLogForwarderSpec{}, "foobar", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)

				collector = podSpec.Containers[0]
				Expect(podSpec.Volumes).To(ContainElement(expectedPodSpecMetricsVol))
				Expect(podSpec.ServiceAccountName).To(Equal("custom-clf"))
				Expect(podSpec.Volumes).To(IncludeVolume(v1.Volume{
					Name: constants.VolumeNameTrustedCA,
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: factory.ResourceNames.CaTrustBundle,
							},
							Items: []v1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}))
				Expect(podSpec.Containers[0].VolumeMounts).To(IncludeVolumeMount(v1.VolumeMount{
					Name:      constants.VolumeNameTrustedCA,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}))
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
		collectionSpec = obs.CollectorSpec{
			Resources: &v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("120Gi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("100Gi"),
					v1.ResourceCPU:    resource.MustParse("500m"),
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

	BeforeEach(func() {
		factory = &Factory{
			ImageName: constants.VectorName,
			Visit:     vector.CollectorVisitor,
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

var _ = Describe("Factory#NewPodSpec Add Cloudwatch STS Resources", func() {

	const (
		tokenSecretName = "mysecret"
	)
	var (
		factory   *Factory
		pipelines = []obs.PipelineSpec{
			{
				Name:       "cw-forward",
				InputRefs:  []string{string(obs.InputTypeInfrastructure)},
				OutputRefs: []string{"cw"},
			},
		}
		outputs = []obs.OutputSpec{
			{
				Type: obs.OutputTypeCloudwatch,
				Name: "cw",
				Cloudwatch: &obs.Cloudwatch{
					Region:  "us-east-77",
					GroupBy: obs.LogGroupByNamespaceName,
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeIAMRole,
						IAMRole: &obs.CloudwatchIAMRole{
							RoleARN: &obs.SecretKey{
								Key: "credentials",
								Secret: &v1.LocalObjectReference{
									Name: "cw",
								},
							},
							Token: &obs.BearerToken{
								From: obs.BearerTokenFromSecret,
								Secret: &obs.BearerTokenSecretKey{
									Key:  constants.TokenKey,
									Name: tokenSecretName,
								},
							},
						},
					},
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
			tokenSecretName: {
				Data: map[string][]byte{
					"token": []byte("abcdef"),
				},
			},
		}
	)
	BeforeEach(func() {
		factory = &Factory{
			ImageName:     constants.VectorName,
			Visit:         vector.CollectorVisitor,
			Secrets:       secrets,
			ResourceNames: coreFactory.ResourceNames(*obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName, runtime.Initialize)),
		}
	})
	Context("when collector has a secret containing a credentials key", func() {

		It("should add the AWS web identity env vars in the container", func() {
			podSpec := *factory.NewPodSpec(nil, obs.ClusterLogForwarderSpec{
				Outputs:   outputs,
				Pipelines: pipelines,
			}, "1234", tls.GetClusterTLSProfileSpec(nil), nil, constants.OpenshiftNS)
			collector := podSpec.Containers[0]

			Expect(collector.Env).To(IncludeEnvVar(v1.EnvVar{
				Name:  constants.AWSRegionEnvVarKey,
				Value: outputs[0].Cloudwatch.Region,
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
				Value: common.SecretPath(tokenSecretName, constants.TokenKey),
			}))
		})
	})
})

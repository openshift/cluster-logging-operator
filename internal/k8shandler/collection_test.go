package k8shandler

import (
	"context"

	"github.com/openshift/cluster-logging-operator/internal/migrations"

	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	securityv1 "github.com/openshift/api/security/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconciling", func() {
	defer GinkgoRecover()

	_ = loggingv1.SchemeBuilder.AddToScheme(scheme.Scheme)
	_ = monitoringv1.AddToScheme(scheme.Scheme)
	_ = securityv1.AddToScheme(scheme.Scheme)
	_ = configv1.AddToScheme(scheme.Scheme)

	var (
		cluster = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLoggingSpec{
				ManagementState: loggingv1.ManagementStateManaged,
				LogStore: &loggingv1.LogStoreSpec{
					Type: loggingv1.LogStoreTypeElasticsearch,
				},
				Collection: &loggingv1.CollectionSpec{
					Type:          loggingv1.LogCollectionTypeVector,
					CollectorSpec: loggingv1.CollectorSpec{},
				},
			},
		}

		collectorSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.CollectorName,
				Namespace: cluster.GetNamespace(),
			},
		}
		collectorCABundle = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.CollectorTrustedCAName,
				Namespace: cluster.GetNamespace(),
				Labels: map[string]string{
					constants.InjectTrustedCABundleLabel: "true",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "ClusterLogging",
						Name:       "instance",
						APIVersion: "logging.openshift.io/v1",
						Controller: utils.GetPtr(true),
					},
				},
			},
			Data: map[string]string{
				constants.TrustedCABundleKey: "",
			},
		}
		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   cluster.Namespace,
			},
		}
		extras = map[string]bool{}
	)

	Describe("Collection", func() {
		var (
			client         cli.Client
			clusterRequest *ClusterLoggingRequest
			spec           loggingv1.ClusterLogForwarderSpec
		)

		Context("when cluster proxy present", func() {
			var (
				customCABundle = `
                  -----BEGIN CERTIFICATE-----
                  <PEM_ENCODED_CERT1>
                  -----END CERTIFICATE-------
                  -----BEGIN CERTIFICATE-----
                  <PEM_ENCODED_CERT2>
                  -----END CERTIFICATE-------
                `
				trustedCABundleVolume = corev1.Volume{
					Name: constants.CollectorTrustedCAName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: constants.CollectorTrustedCAName,
							},
							Items: []corev1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}
				trustedCABundleVolumeMount = corev1.VolumeMount{
					Name:      constants.CollectorTrustedCAName,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}
			)
			BeforeEach(func() {
				cluster.TypeMeta.SetGroupVersionKind(loggingv1.GroupVersion.WithKind("ClusterLogging"))
				client = fake.NewFakeClient( //nolint
					cluster,
					collectorSecret,
					collectorCABundle,
					namespace,
				)
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Reader:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName),
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
					ResourceOwner: utils.AsOwner(cluster),
					isDaemonset:   true,
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras, _ = migrations.MigrateClusterLogForwarder(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret, clusterRequest.ResourceNames.ServiceAccountTokenSecret)
				clusterRequest.Forwarder.Spec = spec
			})

			It("should use the injected custom CA bundle for the collector as daemonset", func() {
				// Reconcile w/o custom CA bundle
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				// Inject custom CA bundle into collector config map
				injectedCABundle := collectorCABundle.DeepCopy()
				injectedCABundle.Data[constants.TrustedCABundleKey] = customCABundle
				Expect(client.Update(context.TODO(), injectedCABundle)).Should(Succeed())

				// Reconcile with injected custom CA bundle
				Expect(clusterRequest.CreateOrUpdateCollection()).Should(Succeed())

				key := types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				ds := &appsv1.DaemonSet{}
				Expect(client.Get(context.TODO(), key, ds)).Should(Succeed())

				bundleVar, found := utils.GetEnvVar(common.TrustedCABundleHashName, ds.Spec.Template.Spec.Containers[0].Env)
				Expect(found).To(BeTrue(), "Exp. the trusted bundle CA hash to be added to the collector container")
				Expect(collector.CalcTrustedCAHashValue(injectedCABundle)).To(Equal(bundleVar.Value))
				Expect(ds.Spec.Template.Spec.Volumes).To(ContainElement(trustedCABundleVolume))
				Expect(ds.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(trustedCABundleVolumeMount))
			})

			It("should use the injected custom CA bundle for the collector as deployment", func() {
				clusterRequest.isDaemonset = false
				// Reconcile w/o custom CA bundle
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				// Inject custom CA bundle into collector config map
				injectedCABundle := collectorCABundle.DeepCopy()
				injectedCABundle.Data[constants.TrustedCABundleKey] = customCABundle
				Expect(client.Update(context.TODO(), injectedCABundle)).Should(Succeed())

				// Reconcile with injected custom CA bundle
				Expect(clusterRequest.CreateOrUpdateCollection()).Should(Succeed())

				key := types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				dpl := &appsv1.Deployment{}
				Expect(client.Get(context.TODO(), key, dpl)).Should(Succeed())

				bundleVar, found := utils.GetEnvVar(common.TrustedCABundleHashName, dpl.Spec.Template.Spec.Containers[0].Env)
				Expect(found).To(BeTrue(), "Exp. the trusted bundle CA hash to be added to the collector container")
				Expect(collector.CalcTrustedCAHashValue(injectedCABundle)).To(Equal(bundleVar.Value))
				Expect(dpl.Spec.Template.Spec.Volumes).To(ContainElement(trustedCABundleVolume))
				Expect(dpl.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(trustedCABundleVolumeMount))
			})
		})
		Context("when cluster proxy is not present", func() {
			var (
				trustedCABundleVolume = corev1.Volume{
					Name: constants.CollectorTrustedCAName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: constants.CollectorTrustedCAName,
							},
							Items: []corev1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}
				trustedCABundleVolumeMount = corev1.VolumeMount{
					Name:      constants.CollectorTrustedCAName,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}
			)
			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					collectorSecret,
					collectorCABundle,
					namespace,
				)
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Reader:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     runtime.NewClusterLogForwarder(constants.OpenshiftNS, "bar"),
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
					ResourceOwner: utils.AsOwner(cluster),
					isDaemonset:   true,
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras, _ = migrations.MigrateClusterLogForwarder(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret, clusterRequest.ResourceNames.ServiceAccountTokenSecret)
				clusterRequest.Forwarder.Spec = spec
			})

			//https://issues.redhat.com/browse/LOG-1859
			It("should continue to reconcile without error as daemonset", func() {
				Expect(clusterRequest.CreateOrUpdateCollection()).Should(Succeed())

				key := types.NamespacedName{Name: constants.CollectorTrustedCAName, Namespace: cluster.GetNamespace()}
				CABundle := &corev1.ConfigMap{}
				Expect(client.Get(context.TODO(), key, CABundle)).Should(Succeed())
				Expect(collectorCABundle.Data).To(Equal(CABundle.Data))

				key = types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				ds := &appsv1.DaemonSet{}
				Expect(client.Get(context.TODO(), key, ds)).Should(Succeed())

				bundleVar, found := utils.GetEnvVar(common.TrustedCABundleHashName, ds.Spec.Template.Spec.Containers[0].Env)
				Expect(found).To(BeTrue(), "Exp. the trusted bundle CA hash to be added to the collector container")
				Expect(bundleVar.Value).To(BeEmpty())
				Expect(ds.Spec.Template.Spec.Volumes).To(Not(ContainElement(trustedCABundleVolume)))
				Expect(ds.Spec.Template.Spec.Containers[0].VolumeMounts).To(Not(ContainElement(trustedCABundleVolumeMount)))
			})

			It("should continue to reconcile without error as deployment", func() {
				clusterRequest.isDaemonset = false
				Expect(clusterRequest.CreateOrUpdateCollection()).Should(Succeed())
				key := types.NamespacedName{Name: constants.CollectorTrustedCAName, Namespace: cluster.GetNamespace()}
				CABundle := &corev1.ConfigMap{}
				Expect(client.Get(context.TODO(), key, CABundle)).Should(Succeed())
				Expect(collectorCABundle.Data).To(Equal(CABundle.Data))

				key = types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				dpl := &appsv1.Deployment{}
				Expect(client.Get(context.TODO(), key, dpl)).Should(Succeed())

				bundleVar, found := utils.GetEnvVar(common.TrustedCABundleHashName, dpl.Spec.Template.Spec.Containers[0].Env)
				Expect(found).To(BeTrue(), "Exp. the trusted bundle CA hash to be added to the collector container")
				Expect(bundleVar.Value).To(BeEmpty())
				Expect(dpl.Spec.Template.Spec.Volumes).To(Not(ContainElement(trustedCABundleVolume)))
				Expect(dpl.Spec.Template.Spec.Containers[0].VolumeMounts).To(Not(ContainElement(trustedCABundleVolumeMount)))
			})
		})

		Context("when removing collector", func() {
			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					collectorSecret,
					collectorCABundle,
					namespace,
				)

				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Reader:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     &loggingv1.ClusterLogForwarder{},
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras, _ = migrations.MigrateClusterLogForwarder(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret, clusterRequest.ResourceNames.ServiceAccountTokenSecret)
				clusterRequest.Forwarder.Spec = spec
			})
		})

		Context("when creating a ClusterLogForwarder not named instance", func() {
			customCLFName := "custom-clf"
			vectorSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      customCLFName,
					Namespace: cluster.GetNamespace(),
				},
			}
			vectorCABundle := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      customCLFName + "-trustbundle",
					Namespace: cluster.GetNamespace(),
					Labels: map[string]string{
						constants.InjectTrustedCABundleLabel: "true",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name:       customCLFName,
							Kind:       "ClusterLogForwarder",
							APIVersion: "logging.openshift.io/v1",
							Controller: utils.GetPtr(true),
						},
					},
				},
				Data: map[string]string{
					constants.TrustedCABundleKey: "",
				},
			}

			fwder := runtime.NewClusterLogForwarder(constants.OpenshiftNS, customCLFName)

			var getObject = func(objName string, obj cli.Object) error {
				nsName := types.NamespacedName{Name: objName, Namespace: cluster.GetNamespace()}
				return client.Get(context.TODO(), nsName, obj)
			}

			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					vectorSecret,
					vectorCABundle,
					namespace,
				)
				cluster.Spec.Collection = &loggingv1.CollectionSpec{
					Type: loggingv1.LogCollectionTypeVector,
				}
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Reader:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     fwder,
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, customCLFName)),
					ResourceOwner: utils.AsOwner(fwder),
					isDaemonset:   true,
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras, _ = migrations.MigrateClusterLogForwarder(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret, clusterRequest.ResourceNames.ServiceAccountTokenSecret)
				clusterRequest.Forwarder.Spec = spec
			})
			It("should have appropriately named resources with daemonset", func() {
				cluster.Spec.Collection.Type = loggingv1.LogCollectionTypeVector
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				// Daemonset
				ds := &appsv1.DaemonSet{}
				Expect(getObject(customCLFName, ds)).Should(Succeed())

			})

			It("should have appropriately named resources with deployment", func() {
				clusterRequest.isDaemonset = false
				cluster.Spec.Collection.Type = loggingv1.LogCollectionTypeVector
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				ds := &appsv1.Deployment{}
				Expect(getObject(customCLFName, ds)).Should(Succeed())
			})
		})
	})
})

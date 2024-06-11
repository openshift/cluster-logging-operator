package observability_test

import (
	"context"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/controller/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
	cli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// The intent of these tests is to verify the nuances of the round trip reconciliation process for outcomes such as:
// * ensuring the cluster-wide proxy certs are mounted into the container
// * the collector is deployed as a DaemonSet or Deployment based upon the available inputs
var _ = Describe("Reconciling the Collector", func() {
	defer GinkgoRecover()

	_ = loggingv1.SchemeBuilder.AddToScheme(scheme.Scheme)
	_ = monitoringv1.AddToScheme(scheme.Scheme)
	_ = securityv1.Install(scheme.Scheme)
	_ = configv1.Install(scheme.Scheme)

	const (
		namespaceName = "mylogging"
		secretName    = "mysecrets"
		clfName       = "mycollector"
		clusterID     = "12345"
	)
	var (
		collectorSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespaceName,
			},
		}

		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   namespaceName,
			},
		}
		client            cli.Client
		forwarder         = obsruntime.NewClusterLogForwarder(namespaceName, clfName, runtime.Initialize)
		receiverForwarder = obsruntime.NewClusterLogForwarder(namespaceName, clfName, runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
			clf.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: ""}
			clf.Spec = obs.ClusterLogForwarderSpec{
				Inputs: []obs.InputSpec{
					{
						Name:     "myreceiver",
						Type:     obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{},
					},
				},
			}
		})
	)

	Context("#ReconcileCollector", func() {
		var (
			customCABundle = `
                  -----BEGIN CERTIFICATE-----
                  <PEM_ENCODED_CERT1>
                  -----END CERTIFICATE-------
                  -----BEGIN CERTIFICATE-----
                  <PEM_ENCODED_CERT2>
                  -----END CERTIFICATE-------
                `
			customCABundlerHash, _ = utils.CalculateMD5Hash(customCABundle)
			resourceNames          *factory.ForwarderResourceNames
			trustedCABundleVolume  = corev1.Volume{
				Name: constants.VolumeNameTrustedCA,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "changeme",
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
				Name:      constants.VolumeNameTrustedCA,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			}
			collectorCABundle = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "changeme",
					Namespace: namespaceName,
					Labels: map[string]string{
						constants.InjectTrustedCABundleLabel: "true",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       "ClusterLogForwarder",
							Name:       clfName,
							APIVersion: "observability.openshift.io/v1",
							Controller: utils.GetPtr(true),
						},
					},
				},
				Data: map[string]string{
					constants.TrustedCABundleKey: "",
				},
			}

			beforeEach = func(clf *obs.ClusterLogForwarder) {
				resourceNames = factory.ResourceNames(*forwarder)
				collectorCABundle.Name = resourceNames.CaTrustBundle
				trustedCABundleVolume.VolumeSource.ConfigMap.Name = resourceNames.CaTrustBundle
				client = fake.NewFakeClient( //nolint
					collectorSecret,
					collectorCABundle,
					namespace,
				)
			}
			reconcileCollector = func(clf *obs.ClusterLogForwarder) {
				Expect(observability.ReconcileCollector(client, client, *clf, clusterID, 1*time.Millisecond, 1*time.Millisecond)).Should(Succeed())
			}
			podTemplateSpecFromDeployment = func(obj cli.Object) corev1.PodTemplateSpec {
				d := obj.(*appsv1.Deployment)
				return d.Spec.Template
			}
			podTemplateFromDaemonSet = func(obj cli.Object) corev1.PodTemplateSpec {
				ds := obj.(*appsv1.DaemonSet)
				return ds.Spec.Template
			}
		)
		It("should deploy services for input receivers", func() {
			clf := obsruntime.NewClusterLogForwarder(namespaceName, clfName, runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
				clf.Spec = obs.ClusterLogForwarderSpec{
					Inputs: []obs.InputSpec{
						{
							Name:        string(obs.InputTypeApplication),
							Type:        obs.InputTypeApplication,
							Application: &obs.Application{},
						},
						{
							Name: "mysyslog",
							Type: obs.InputTypeReceiver,
							Receiver: &obs.ReceiverSpec{
								Port: 12345,
								Type: obs.ReceiverTypeSyslog,
							},
						},
						{
							Name: "myhttp",
							Type: obs.InputTypeReceiver,
							Receiver: &obs.ReceiverSpec{
								Port: 12345,
								Type: obs.ReceiverTypeHTTP,
								HTTP: &obs.HTTPReceiver{},
							},
						},
					},
				}
			})
			beforeEach(clf)
			reconcileCollector(clf)

			for _, input := range clf.Spec.Inputs {
				if input.Type == obs.InputTypeReceiver {
					key := types.NamespacedName{Name: resourceNames.GenerateInputServiceName(input.Name), Namespace: namespaceName}
					service := &corev1.Service{}
					Expect(client.Get(context.TODO(), key, service)).Should(Succeed(), "Exp. to create a Service for input:", input.Name)
				}
			}
		})
		DescribeTable("should deploy resources to support metrics collection", func(clf *obs.ClusterLogForwarder) {
			beforeEach(clf)
			reconcileCollector(clf)

			key := types.NamespacedName{Name: clfName, Namespace: namespaceName}
			service := &corev1.Service{}
			Expect(client.Get(context.TODO(), key, service)).Should(Succeed(), "Exp. to create a Service for metrics")

			sm := &monitoringv1.ServiceMonitor{}
			Expect(client.Get(context.TODO(), key, sm)).Should(Succeed(), "Exp. to create a ServiceMonitor for metrics")

		},
			Entry("when deployed as a DaemonSet", forwarder),
			Entry("when deployed as a Deployment", receiverForwarder),
		)

		DescribeTable("when the cluster proxy is present should use the injected custom CA bundle", func(clf *obs.ClusterLogForwarder, obj cli.Object, templateSpec func(obj cli.Object) corev1.PodTemplateSpec) {
			beforeEach(clf)

			// Reconcile w/o custom CA bundle
			reconcileCollector(clf)

			// Inject custom CA bundle into collector config map
			injectedCABundle := collectorCABundle.DeepCopy()
			injectedCABundle.Data[constants.TrustedCABundleKey] = customCABundle
			Expect(client.Update(context.TODO(), injectedCABundle)).Should(Succeed())

			// Reconcile with injected custom CA bundle
			reconcileCollector(clf)

			key := types.NamespacedName{Name: clfName, Namespace: namespaceName}
			Expect(client.Get(context.TODO(), key, obj)).Should(Succeed(), "Exp. to create a deployed collector")
			podTemplateSpec := templateSpec(obj)
			Expect(podTemplateSpec.Spec.Containers[0].Env).To(IncludeEnvVar(corev1.EnvVar{Name: common.TrustedCABundleHashName, Value: customCABundlerHash}), "Exp. the trusted bundle CA hash to be added to the collector container")
			Expect(podTemplateSpec.Spec.Volumes).To(IncludeVolume(trustedCABundleVolume))
			Expect(podTemplateSpec.Spec.Containers[0].VolumeMounts).To(IncludeVolumeMount(trustedCABundleVolumeMount))
		},
			Entry("when deployed as a DaemonSet", forwarder, &appsv1.DaemonSet{}, podTemplateFromDaemonSet),
			Entry("when deployed as a Deployment", receiverForwarder, &appsv1.Deployment{}, podTemplateSpecFromDeployment),
		)
		DescribeTable("when the cluster proxy is not present should not error", func(clf *obs.ClusterLogForwarder, obj cli.Object, templateSpec func(obj cli.Object) corev1.PodTemplateSpec) {
			beforeEach(clf)

			// Reconcile w/o custom CA bundle
			reconcileCollector(clf)

			key := types.NamespacedName{Name: clfName, Namespace: namespaceName}
			Expect(client.Get(context.TODO(), key, obj)).Should(Succeed(), "Exp. to create a deployed collector")
			podTemplateSpec := templateSpec(obj)
			Expect(podTemplateSpec.Spec.Containers[0].Env).To(Not(IncludeEnvVar(corev1.EnvVar{Name: common.TrustedCABundleHashName, Value: customCABundlerHash})), "Exp. the trusted bundle CA hash to not be added to the collector container")
			Expect(podTemplateSpec.Spec.Volumes).To(Not(IncludeVolume(trustedCABundleVolume)))
			Expect(podTemplateSpec.Spec.Containers[0].VolumeMounts).To(Not(IncludeVolumeMount(trustedCABundleVolumeMount)))
		},
			Entry("when deployed as a DaemonSet", forwarder, &appsv1.DaemonSet{}, podTemplateFromDaemonSet),
			Entry("when deployed as a Deployment", receiverForwarder, &appsv1.Deployment{}, podTemplateSpecFromDeployment),
		)
	})
})

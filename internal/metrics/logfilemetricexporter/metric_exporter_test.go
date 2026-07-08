package logfilemetricexporter

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1 "github.com/openshift/api/security/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile LogFileMetricExporter", func() {

	defer GinkgoRecover()
	_ = monitoringv1.AddToScheme(scheme.Scheme)
	_ = securityv1.AddToScheme(scheme.Scheme)

	var (
		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   constants.OpenshiftNS,
			},
		}

		reqClient = fake.NewFakeClient(
			namespace,
		)

		reader = reqClient.(client.Reader)

		lfmeInstance = &loggingv1alpha1.LogFileMetricExporter{}

		// Daemonset
		dsKey      = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		dsInstance = &appsv1.DaemonSet{}
		reqMem1    = resource.MustParse("50Gi")
		reqCPU1    = resource.MustParse("300m")
		reqMem2    = resource.MustParse("15Gi")
		reqCPU2    = resource.MustParse("100m")

		// Service
		serviceKey      = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		serviceInstance = &corev1.Service{}

		// Service Monitor
		serviceMonitorKey = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		smInstance        = &monitoringv1.ServiceMonitor{}
	)

	It("Should reconcile successfully a daemonset, service, and service monitor", func() {

		runtime.Initialize(lfmeInstance, constants.OpenshiftNS, constants.SingletonName)

		lfmeInstance.Spec = loggingv1alpha1.LogFileMetricExporterSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    reqCPU2,
					corev1.ResourceMemory: reqMem2,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    reqCPU1,
					corev1.ResourceMemory: reqMem1,
				},
			},
		}

		// Reconcile the LogFileMetricExporter
		Expect(Reconcile(lfmeInstance, reqClient, reader, utils.AsOwner(lfmeInstance))).To(Succeed())

		// Daemonset
		// Get and check the daemonset
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(Not(BeNil()))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(Not(BeNil()))

		// Check resource limits
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().Cmp(reqCPU2)).To(Equal(0))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Cmp(reqMem2)).To(Equal(0))

		// Check request limits
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().Cmp(reqCPU1)).To(Equal(0))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Cmp(reqMem1)).To(Equal(0))

		// Service
		// Get and check the service
		Expect(reqClient.Get(context.TODO(), serviceKey, serviceInstance)).Should(Succeed())

		Expect(serviceInstance.Name).To(Equal(constants.LogfilesmetricexporterName))
		Expect(serviceInstance.Spec.Ports).ToNot(BeEmpty(), "Exp. to have spec.Ports")

		Expect(serviceInstance.Spec.Ports[0].Port).
			To(Equal(exporterPort), fmt.Sprintf("Exp service port of: %v", exporterPort))

		Expect(serviceInstance.Annotations[constants.AnnotationServingCertSecretName]).
			To(Equal(ExporterMetricsSecretName))

		// ServiceMonitor (full profile)
		Expect(reqClient.Get(context.TODO(), serviceMonitorKey, smInstance)).Should(Succeed())

		Expect(smInstance.Name).To(Equal(constants.LogfilesmetricexporterName))
		Expect(smInstance.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileFull))

		expJobLabel := fmt.Sprintf("monitor-%s", constants.LogfilesmetricexporterName)
		Expect(smInstance.Spec.JobLabel).To(Equal(expJobLabel))
		Expect(smInstance.Spec.Endpoints).ToNot(BeEmpty())
		Expect(smInstance.Spec.Endpoints[0].Port).To(Equal(constants.MetricsPortName))

		svcURL := fmt.Sprintf("%s.openshift-logging.svc", constants.LogfilesmetricexporterName)
		Expect(smInstance.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).
			To(Equal(svcURL))

		Expect(smInstance.Spec.Endpoints[0].BearerTokenFile).
			To(Equal("/var/run/secrets/kubernetes.io/serviceaccount/token"))
		Expect(smInstance.Spec.Endpoints[0].MetricRelabelConfigs).To(HaveLen(1), "full profile should only have the rename rule")

		// ServiceMonitor (minimal profile)
		minimalName := constants.MetricsCollectionProfileMinimal + "-" + constants.LogfilesmetricexporterName
		minimalSM := &monitoringv1.ServiceMonitor{}
		Expect(reqClient.Get(context.TODO(), types.NamespacedName{Name: minimalName, Namespace: namespace.Name}, minimalSM)).Should(Succeed())
		Expect(minimalSM.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileMinimal))
		Expect(minimalSM.Spec.Endpoints).ToNot(BeEmpty())
		Expect(minimalSM.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).To(Equal(svcURL))
		Expect(minimalSM.Spec.Endpoints[0].MetricRelabelConfigs).To(HaveLen(2), "LFME minimal profile should have rename + keep")
		Expect(string(minimalSM.Spec.Endpoints[0].MetricRelabelConfigs[1].Action)).To(Equal("keep"))

		// ServiceMonitor (telemetry profile)
		telemetryName := constants.MetricsCollectionProfileTelemetry + "-" + constants.LogfilesmetricexporterName
		telemetrySM := &monitoringv1.ServiceMonitor{}
		Expect(reqClient.Get(context.TODO(), types.NamespacedName{Name: telemetryName, Namespace: namespace.Name}, telemetrySM)).Should(Succeed())
		Expect(telemetrySM.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileTelemetry))
		Expect(telemetrySM.Spec.Endpoints).ToNot(BeEmpty())
		Expect(telemetrySM.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).To(Equal(svcURL))
		Expect(telemetrySM.Spec.Endpoints[0].MetricRelabelConfigs).To(HaveLen(2), "LFME telemetry profile should have rename + keep")
		Expect(string(telemetrySM.Spec.Endpoints[0].MetricRelabelConfigs[1].Action)).To(Equal("keep"))

		// Metrics Auth RBAC
		// Verify the metrics auth ClusterRoleBinding exists and references system:auth-delegator
		metricsAuthBinding := &rbacv1.ClusterRoleBinding{}
		expectedMetricsAuthName := fmt.Sprintf("cluster-logging-%s-%s-metrics-auth", namespace.Name, constants.LogfilesmetricexporterName)
		Expect(reqClient.Get(context.TODO(), types.NamespacedName{Name: expectedMetricsAuthName}, metricsAuthBinding)).Should(Succeed())
		Expect(metricsAuthBinding.RoleRef.Name).To(Equal("system:auth-delegator"))
		Expect(metricsAuthBinding.Subjects).To(HaveLen(1))
		Expect(metricsAuthBinding.Subjects[0].Name).To(Equal(constants.LogfilesmetricexporterName))
	})

	Context("when the logfilemetricexporter NetworkPolicy is reconciled", func() {
		var (
			networkPolicyKey      types.NamespacedName
			networkPolicyInstance *networkingv1.NetworkPolicy
		)

		BeforeEach(func() {
			runtime.Initialize(lfmeInstance, constants.OpenshiftNS, constants.SingletonName)
			// NetworkPolicy naming convention: "lfme-" + constants.LogfilesmetricexporterName
			networkPolicyKey = types.NamespacedName{
				Name:      "lfme-" + constants.LogfilesmetricexporterName,
				Namespace: constants.OpenshiftNS,
			}
			networkPolicyInstance = &networkingv1.NetworkPolicy{}
		})

		It("should not create a NetworkPolicy when not specified", func() {
			// NetworkPolicy should be nil (default)
			Expect(lfmeInstance.Spec.NetworkPolicy).To(BeNil())

			// Reconcile the LogFileMetricExporter
			Expect(Reconcile(lfmeInstance, reqClient, reader, utils.AsOwner(lfmeInstance))).To(Succeed())

			// Verify NetworkPolicy was not created
			err := reqClient.Get(context.TODO(), networkPolicyKey, networkPolicyInstance)
			Expect(errors.IsNotFound(err)).To(BeTrue(), "NetworkPolicy should not exist when not specified")
		})

		Context("when NetworkPolicy is specified", func() {
			It("should successfully create and configure the NetworkPolicy for AllowIngressMetrics ruleset", func() {
				// Configure NetworkPolicy with AllowIngressMetrics ruleset
				lfmeInstance.Spec.NetworkPolicy = &loggingv1alpha1.NetworkPolicy{
					RuleSet: loggingv1alpha1.NetworkPolicyRuleSetTypeAllowIngressMetrics,
				}

				// Reconcile the LogFileMetricExporter
				Expect(Reconcile(lfmeInstance, reqClient, reader, utils.AsOwner(lfmeInstance))).To(Succeed())

				// Get and verify the NetworkPolicy was created
				Expect(reqClient.Get(context.TODO(), networkPolicyKey, networkPolicyInstance)).Should(Succeed())

				// Verify basic properties
				Expect(networkPolicyInstance.Name).To(Equal("lfme-" + constants.LogfilesmetricexporterName))
				Expect(networkPolicyInstance.Namespace).To(Equal(constants.OpenshiftNS))

				// Verify owner reference is set
				Expect(networkPolicyInstance.OwnerReferences).To(HaveLen(1))
				Expect(networkPolicyInstance.OwnerReferences[0].Name).To(Equal(lfmeInstance.Name))

				// Verify pod selector matches LFME pods
				expectedPodSelector := runtime.Selectors(lfmeInstance.Name, constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName)
				Expect(networkPolicyInstance.Spec.PodSelector.MatchLabels).To(Equal(expectedPodSelector))

				// Verify policy types include Ingress and Egress (AllowIngressMetrics ruleset)
				expectedPolicyTypes := []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress}
				Expect(networkPolicyInstance.Spec.PolicyTypes).To(ConsistOf(expectedPolicyTypes))

				// Verify ingress rules allow only the named metrics port
				Expect(networkPolicyInstance.Spec.Ingress).To(HaveLen(1))
				Expect(networkPolicyInstance.Spec.Ingress[0].Ports).To(HaveLen(1))

				expectedPort := networkingv1.NetworkPolicyPort{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &[]intstr.IntOrString{{Type: intstr.Int, IntVal: constants.LogfilesmetricexporterPort}}[0],
				}
				Expect(networkPolicyInstance.Spec.Ingress[0].Ports[0]).To(Equal(expectedPort))

				// Verify egress rules allow DNS and KubeAPI (required for secure metrics token validation)
				Expect(networkPolicyInstance.Spec.Egress).To(HaveLen(1))
				Expect(networkPolicyInstance.Spec.Egress[0].Ports).To(ConsistOf(
					networkingv1.NetworkPolicyPort{
						Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.Int, IntVal: factory.KubeAPIPort}}[0],
					},
					networkingv1.NetworkPolicyPort{
						Protocol: &[]corev1.Protocol{corev1.ProtocolUDP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.String, StrVal: factory.DNSPortName}}[0],
					},
				))

				// Verify common labels are set
				expectedLabels := map[string]string{
					constants.LabelK8sName:      constants.LogfilesmetricexporterName,
					constants.LabelK8sInstance:  lfmeInstance.Name,
					constants.LabelK8sComponent: constants.LogfilesmetricexporterName,
				}
				for key, expectedValue := range expectedLabels {
					Expect(networkPolicyInstance.Labels).To(HaveKeyWithValue(key, expectedValue))
				}
			})
		})
	})
})

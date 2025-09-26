package network

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile NetworkPolicy", func() {

	defer GinkgoRecover()

	var (
		owner          metav1.OwnerReference
		policyInstance *networkingv1.NetworkPolicy
		policyKey      types.NamespacedName
		policyName     string
		componentName  string
		commonLabels   func(o runtime.Object)

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
		instanceName = "test-instance"
		protocolTCP  = corev1.ProtocolTCP
	)

	Context("when the collector NetworkPolicy is reconciled", func() {
		BeforeEach(func() {
			policyName = "collector-test-network-policy"
			componentName = constants.CollectorName
			commonLabels = func(o runtime.Object) {
				runtime.SetCommonLabels(o, constants.VectorName, instanceName, componentName)
			}
			owner = metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "ClusterLogForwarder",
				Name:       instanceName,
			}
			policyInstance = &networkingv1.NetworkPolicy{}
			policyKey = types.NamespacedName{Name: policyName, Namespace: namespace.Name}
		})

		It("should successfully reconcile the network policy", func() {
			// Reconcile the network policy
			Expect(ReconcileClusterLogForwarderNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				"",
				owner,
				commonLabels)).To(Succeed())

			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify basic properties
			Expect(policyInstance.Name).To(Equal(policyName))
			Expect(policyInstance.Namespace).To(Equal(constants.OpenshiftNS))

			// Verify owner reference is set
			Expect(policyInstance.OwnerReferences).To(HaveLen(1))
			Expect(policyInstance.OwnerReferences[0].Name).To(Equal(instanceName))

			// Verify pod selector matches collector pods
			expectedPodSelector := runtime.Selectors(instanceName, componentName, constants.VectorName)
			Expect(policyInstance.Spec.PodSelector.MatchLabels).To(Equal(expectedPodSelector))

			// Verify policy types include both Ingress and Egress
			expectedPolicyTypes := []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			}
			Expect(policyInstance.Spec.PolicyTypes).To(Equal(expectedPolicyTypes))

			expectedIngressRules := []networkingv1.NetworkPolicyIngressRule{
				{},
			}
			// Verify ingress rules allow all traffic (empty rule)
			Expect(policyInstance.Spec.Ingress).To(Equal(expectedIngressRules))

			// Verify egress rules allow all traffic (empty rule)
			expectedEgressRules := []networkingv1.NetworkPolicyEgressRule{
				{},
			}
			Expect(policyInstance.Spec.Egress).To(Equal(expectedEgressRules))

			// Verify common labels are applied
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sName))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sInstance))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sComponent))
			Expect(policyInstance.Labels[constants.LabelK8sName]).To(Equal(constants.VectorName))
			Expect(policyInstance.Labels[constants.LabelK8sInstance]).To(Equal(instanceName))
			Expect(policyInstance.Labels[constants.LabelK8sComponent]).To(Equal(componentName))
		})
	})
	Context("when the logfilemetricexporter NetworkPolicy is reconciled", func() {
		BeforeEach(func() {
			policyName = "lfme-test-network-policy"
			componentName = constants.LogfilesmetricexporterName
			commonLabels = func(o runtime.Object) {
				runtime.SetCommonLabels(o, constants.LogfilesmetricexporterName, instanceName, componentName)
			}
			owner = metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "LogFileMetricExporter",
				Name:       instanceName,
			}
			policyInstance = &networkingv1.NetworkPolicy{}
			policyKey = types.NamespacedName{Name: policyName, Namespace: namespace.Name}
		})

		It("should successfully reconcile the network policy", func() {
			// Reconcile the network policy
			Expect(ReconcileLogFileMetricsExporterNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				loggingv1alpha1.NetworkPolicyRuleSetTypeAllowIngressMetrics,
				owner,
				commonLabels)).To(Succeed())
			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify basic properties
			Expect(policyInstance.Name).To(Equal(policyName))
			Expect(policyInstance.Namespace).To(Equal(constants.OpenshiftNS))

			// Verify owner reference is set
			Expect(policyInstance.OwnerReferences).To(HaveLen(1))
			Expect(policyInstance.OwnerReferences[0].Name).To(Equal(instanceName))

			// Verify pod selector matches LFME pods
			expectedPodSelector := runtime.Selectors(instanceName, componentName, constants.LogfilesmetricexporterName)
			Expect(policyInstance.Spec.PodSelector.MatchLabels).To(Equal(expectedPodSelector))

			// Verify policy types includes Egress and Ingress
			expectedPolicyTypes := []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			}
			Expect(policyInstance.Spec.PolicyTypes).To(Equal(expectedPolicyTypes))

			// Verify ingress rules allow only the named metrics port
			expectedIngressRules := []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: &protocolTCP,
							Port:     &intstr.IntOrString{Type: intstr.String, StrVal: constants.MetricsPortName},
						},
					},
				},
			}
			Expect(policyInstance.Spec.Ingress).To(Equal(expectedIngressRules))

			// Verify egress rules are not present to deny all egress traffic
			Expect(policyInstance.Spec.Egress).To(BeNil())

			// Verify common labels are applied
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sName))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sInstance))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sComponent))
			Expect(policyInstance.Labels[constants.LabelK8sName]).To(Equal(constants.LogfilesmetricexporterName))
			Expect(policyInstance.Labels[constants.LabelK8sInstance]).To(Equal(instanceName))
			Expect(policyInstance.Labels[constants.LabelK8sComponent]).To(Equal(componentName))
		})
	})

})

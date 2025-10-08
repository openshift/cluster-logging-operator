package factory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/version"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("#NewNetworkPolicy", func() {

	var (
		namespace    = "openshift-logging"
		policyName   = "policy-name"
		instanceName = "instance-name"
		commonLabels = func(o runtime.Object) {
			runtime.SetCommonLabels(o, constants.VectorName, instanceName, constants.CollectorName)
		}
	)

	Context("Common properties", func() {
		var np *networkingv1.NetworkPolicy

		BeforeEach(func() {
			np = NewNetworkPolicyWithProtocolPorts(namespace, policyName, instanceName, constants.CollectorName, "", nil, nil, commonLabels)
		})

		It("should set name and namespace correctly", func() {
			Expect(np.Name).To(Equal(policyName))
			Expect(np.Namespace).To(Equal(namespace))
		})

		It("should set labels correctly", func() {
			expectedLabels := map[string]string{
				constants.LabelK8sName:      constants.VectorName,
				constants.LabelK8sInstance:  instanceName,
				constants.LabelK8sComponent: constants.CollectorName,
				constants.LabelK8sPartOf:    constants.ClusterLogging,
				constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
				constants.LabelK8sVersion:   version.Version,
			}

			Expect(np.Labels).To(Equal(expectedLabels))
		})

		It("should set pod selector to match component pods", func() {
			expectedPodSelector := metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.LabelK8sName:      constants.VectorName,
					constants.LabelK8sInstance:  instanceName,
					constants.LabelK8sComponent: constants.CollectorName,
					constants.LabelK8sPartOf:    constants.ClusterLogging,
					constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
				},
			}
			Expect(np.Spec.PodSelector).To(Equal(expectedPodSelector))
		})
	})

	DescribeTable("Ruleset configuration",
		func(ruleSet string, ingressPorts []int32, egressPorts []PortProtocol, expectedPolicyTypes []networkingv1.PolicyType, expectedIngressRules []networkingv1.NetworkPolicyIngressRule, expectedEgressRules []networkingv1.NetworkPolicyEgressRule) {
			np := NewNetworkPolicyWithProtocolPorts(namespace, policyName, instanceName, constants.CollectorName, ruleSet, egressPorts, ingressPorts, commonLabels)

			// Verify policy types are set correctly based on the ruleset
			Expect(np.Spec.PolicyTypes).To(ConsistOf(expectedPolicyTypes))

			// Verify ingress rules match expectations
			Expect(np.Spec.Ingress).To(Equal(expectedIngressRules))

			// Verify egress rules match expectations
			Expect(np.Spec.Egress).To(Equal(expectedEgressRules))
		},
		Entry("with default ruleset (allow all ingress and egress)",
			"",
			nil,
			nil,
			[]networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
			[]networkingv1.NetworkPolicyIngressRule{{}}, // Empty rule allows all ingress
			[]networkingv1.NetworkPolicyEgressRule{{}},  // Empty rule allows all egress
		),
		Entry("with AllowIngressMetrics ruleset",
			string(loggingv1alpha1.NetworkPolicyRuleSetTypeAllowIngressMetrics),
			nil,
			nil,
			[]networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
			[]networkingv1.NetworkPolicyIngressRule{{
				Ports: []networkingv1.NetworkPolicyPort{{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &[]intstr.IntOrString{{Type: intstr.String, StrVal: constants.MetricsPortName}}[0],
				}},
			}},
			nil, // No egress rules
		),
		Entry("with RestrictIngressEgress ruleset",
			string(obsv1.NetworkPolicyRuleSetTypeRestrictIngressEgress),
			[]int32{5000},
			[]PortProtocol{{Port: 8080, Protocol: corev1.ProtocolTCP}, {Port: 5140, Protocol: corev1.ProtocolTCP}, {Port: 9200, Protocol: corev1.ProtocolTCP}},
			[]networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
			[]networkingv1.NetworkPolicyIngressRule{{
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.String, StrVal: constants.MetricsPortName}}[0],
					},
					{
						Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.Int, IntVal: 5000}}[0],
					},
				},
			}},
			[]networkingv1.NetworkPolicyEgressRule{{
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.Int, IntVal: 8080}}[0],
					},
					{
						Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.Int, IntVal: 5140}}[0],
					},
					{
						Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						Port:     &[]intstr.IntOrString{{Type: intstr.Int, IntVal: 9200}}[0],
					},
				},
			}},
		),
	)
})

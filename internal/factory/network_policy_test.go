package factory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/version"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			np = NewNetworkPolicy(namespace, policyName, instanceName, constants.CollectorName, []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress}, commonLabels)
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

	DescribeTable("Policy types configuration",
		func(policyTypes []networkingv1.PolicyType, expectIngress bool, expectEgress bool) {
			np := NewNetworkPolicy(namespace, policyName, instanceName, constants.CollectorName, policyTypes, commonLabels)

			// Verify policy types are set correctly
			Expect(np.Spec.PolicyTypes).To(Equal(policyTypes))

			// Verify ingress rules based on expectation
			if expectIngress {
				Expect(np.Spec.Ingress).To(HaveLen(1))
				Expect(np.Spec.Ingress[0]).To(Equal(networkingv1.NetworkPolicyIngressRule{}))
			} else {
				Expect(np.Spec.Ingress).To(BeNil())
			}

			// Verify egress rules based on expectation
			if expectEgress {
				Expect(np.Spec.Egress).To(HaveLen(1))
				Expect(np.Spec.Egress[0]).To(Equal(networkingv1.NetworkPolicyEgressRule{}))
			} else {
				Expect(np.Spec.Egress).To(BeNil())
			}
		},
		Entry("with Ingress policy type only",
			[]networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			true,  
			false,
		),
		Entry("with Egress policy type only",
			[]networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			false,
			true,  
		),
		Entry("with both Ingress and Egress policy types",
			[]networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
			true, 
			true, 
		),
		Entry("with empty policy types array",
			[]networkingv1.PolicyType{},
			false, 
			false, 
		),
	)
})

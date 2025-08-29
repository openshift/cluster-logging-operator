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
		np           *networkingv1.NetworkPolicy
		namespace    = "openshift-logging"
		policyName   = "policy-name"
		instanceName = "instance-name"
		commonLabels = func(o runtime.Object) {
			runtime.SetCommonLabels(o, constants.VectorName, instanceName, constants.CollectorName)
		}
	)

	BeforeEach(func() {
		np = NewNetworkPolicy(namespace, policyName, instanceName, commonLabels)
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

	It("should set pod selector to match collector pods", func() {
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

	It("should include both Ingress and Egress policy types", func() {
		expectedPolicyTypes := []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		}
		Expect(np.Spec.PolicyTypes).To(Equal(expectedPolicyTypes))
	})

	It("should have ingress rules that allow all traffic", func() {
		Expect(np.Spec.Ingress).To(HaveLen(1))
		Expect(np.Spec.Ingress[0]).To(Equal(networkingv1.NetworkPolicyIngressRule{}))
	})

	It("should have egress rules that allow all traffic", func() {
		Expect(np.Spec.Egress).To(HaveLen(1))
		Expect(np.Spec.Egress[0]).To(Equal(networkingv1.NetworkPolicyEgressRule{}))
	})
})

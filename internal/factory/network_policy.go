package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewNetworkPolicy creates a NetworkPolicy for the given component.
// It configures the policy to allow all ingress and egress traffic for the component pods.
func NewNetworkPolicy(namespace, policyName, instanceName, component string, policyTypes []networkingv1.PolicyType, visitors ...func(o runtime.Object)) *networkingv1.NetworkPolicy {
	// Create the base NetworkPolicy
	np := runtime.NewNetworkPolicy(namespace, policyName, visitors...)

	// Set up pod selector to match the component pods for this instance
	podSelector := runtime.Selectors(instanceName, component, np.Labels[constants.LabelK8sName])

	// Configure the NetworkPolicy spec
	np.Spec = networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: podSelector,
		},
		PolicyTypes: policyTypes,
	}

	for _, policyType := range policyTypes {
		switch policyType {
		case networkingv1.PolicyTypeIngress:
			np.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{
				{}, // Empty rule allows all ingress traffic for the component pods
			}
		case networkingv1.PolicyTypeEgress:
			np.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{
				{}, // Empty rule allows all egress traffic for the component pods	
			}
		}
	}

	return np
}

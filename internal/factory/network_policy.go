package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewNetworkPolicy creates a NetworkPolicy for clfs.
// It configures the policy to allow all ingress and egress traffic for the collector pods.
func NewNetworkPolicy(namespace, policyName, instanceName string, visitors ...func(o runtime.Object)) *networkingv1.NetworkPolicy {
	// Create the base NetworkPolicy
	np := runtime.NewNetworkPolicy(namespace, policyName, visitors...)

	// Set up pod selector to match collector pods for this instance
	podSelector := runtime.Selectors(instanceName, constants.CollectorName, constants.VectorName)

	// Configure the NetworkPolicy spec
	np.Spec = networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: podSelector,
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		// Allow all ingress traffic
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{}, // Empty rule allows all ingress
		},
		// Allow all egress traffic
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{}, // Empty rule allows all egress
		},
	}

	return np
}

package factory

import (
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// NewNetworkPolicy creates a NetworkPolicy for the given component.
// It configures the policy to allow all ingress and egress traffic for the component pods.
func NewNetworkPolicy(namespace, policyName, instanceName, component, policyRuleSet string, visitors ...func(o runtime.Object)) *networkingv1.NetworkPolicy {
	// Create the base NetworkPolicy
	np := runtime.NewNetworkPolicy(namespace, policyName, visitors...)

	// Set up pod selector to match the component pods for this instance
	podSelector := runtime.Selectors(instanceName, component, np.Labels[constants.LabelK8sName])

	npBuilder := runtime.NewNetworkPolicyBuilder(np).WithPodSelector(podSelector)

	// Configure the policy based on the rule set
	// TODO: Add support for other rule sets
	switch policyRuleSet {
	case string(loggingv1alpha1.NetworkPolicyRuleSetTypeAllowIngressMetrics):
		NetworkPolicyTypeAllowIngressMetrics(npBuilder)
	case string(loggingv1alpha1.NetworkPolicyRuleSetTypeAllowAllIngressEgress):
		NetworkPolicyTypeAllowAllIngressEgress(npBuilder)
	default:
		NetworkPolicyTypeAllowAllIngressEgress(npBuilder)
	}

	return np
}

func NetworkPolicyTypeAllowAllIngressEgress(npBuilder *runtime.NetworkPolicyBuilder) *runtime.NetworkPolicyBuilder {
	return npBuilder.
		AllowAllIngress().
		AllowAllEgress()
}

func NetworkPolicyTypeAllowIngressMetrics(npBuilder *runtime.NetworkPolicyBuilder) *runtime.NetworkPolicyBuilder {
	return npBuilder.
		WithEgressPolicyType(). // Adding egress policy type without any rules to deny all egress traffic
		NewIngressRule().
		OnNamedPort(corev1.ProtocolTCP, constants.MetricsPortName).
		End()
}

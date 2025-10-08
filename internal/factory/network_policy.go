package factory

import (
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// PortProtocol represents a port with its associated protocol
type PortProtocol struct {
	Port     int32
	Protocol corev1.Protocol
}

// NewNetworkPolicyWithProtocolPorts creates a NetworkPolicy with protocol-aware port configuration.
func NewNetworkPolicyWithProtocolPorts(namespace, policyName, instanceName, component, policyRuleSet string, egressPorts []PortProtocol, ingressPorts []int32, visitors ...func(o runtime.Object)) *networkingv1.NetworkPolicy {
	// Create the base NetworkPolicy
	np := runtime.NewNetworkPolicy(namespace, policyName, visitors...)

	// Set up pod selector to match the component pods for this instance
	podSelector := runtime.Selectors(instanceName, component, np.Labels[constants.LabelK8sName])

	npBuilder := runtime.NewNetworkPolicyBuilder(np).WithPodSelector(podSelector)

	// Configure the policy based on the rule set
	switch policyRuleSet {
	// allow all ingress and egress traffic
	case string(loggingv1alpha1.NetworkPolicyRuleSetTypeAllowAllIngressEgress):
		NetworkPolicyTypeAllowAllIngressEgress(npBuilder)
	// allow ingress on the metrics port only and deny all egress traffic
	case string(loggingv1alpha1.NetworkPolicyRuleSetTypeAllowIngressMetrics):
		NetworkPolicyTypeAllowIngressMetrics(npBuilder)
	// allow ingress on the specified ports and egress to the specified ports
	case string(obsv1.NetworkPolicyRuleSetTypeRestrictIngressEgress):
		NetworkPolicyTypeRestrictIngressEgressWithProtocols(npBuilder, ingressPorts, egressPorts)
	default:
		NetworkPolicyTypeAllowAllIngressEgress(npBuilder)
	}

	return np
}

// NetworkPolicyTypeAllowAllIngressEgress configures the network policy to allow all ingress and egress traffic.
func NetworkPolicyTypeAllowAllIngressEgress(npBuilder *runtime.NetworkPolicyBuilder) *runtime.NetworkPolicyBuilder {
	return npBuilder.
		AllowAllIngress().
		AllowAllEgress()
}

// NetworkPolicyTypeAllowIngressMetrics configures the network policy to allow ingress on the metrics port only and deny all egress traffic.
func NetworkPolicyTypeAllowIngressMetrics(npBuilder *runtime.NetworkPolicyBuilder) *runtime.NetworkPolicyBuilder {
	return npBuilder.
		WithEgressPolicyType(). // Adding egress policy type without any rules to deny all egress traffic
		NewIngressRule().
		OnNamedPort(corev1.ProtocolTCP, constants.MetricsPortName).
		End()
}

// NetworkPolicyTypeRestrictIngressEgressWithProtocols configures the network policy to restrict ingress and egress traffic
// It allows ingress on specified ports and egress to specified ports with their protocols.
func NetworkPolicyTypeRestrictIngressEgressWithProtocols(npBuilder *runtime.NetworkPolicyBuilder, ingressPorts []int32, egressPorts []PortProtocol) *runtime.NetworkPolicyBuilder {
	// Ingress rules are allowed on the metrics port and all additional spec'd ingress ports
	ingressRule := npBuilder.NewIngressRule().
		OnNamedPort(corev1.ProtocolTCP, constants.MetricsPortName)

	// Add all additional spec'd ingress ports to the same rule (still TCP for inputs)
	for _, port := range ingressPorts {
		ingressRule.OnPort(corev1.ProtocolTCP, port)
	}
	ingressRule.End()

	// Egress rules are allowed on all spec'd egress ports with their detected protocols
	egressRule := npBuilder.NewEgressRule()

	for _, portProtocol := range egressPorts {
		egressRule.OnPort(portProtocol.Protocol, portProtocol.Port)
	}
	egressRule.End()

	return npBuilder
}

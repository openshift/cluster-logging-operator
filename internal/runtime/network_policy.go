package runtime

import (
	"slices"
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/utils/json"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkPolicyBuilder struct {
	NetworkPolicy *networkingv1.NetworkPolicy
}

// NewNetworkPolicyBuilder creates a new NetworkPolicyBuilder with the given NetworkPolicy
func NewNetworkPolicyBuilder(np *networkingv1.NetworkPolicy) *NetworkPolicyBuilder {
	return &NetworkPolicyBuilder{
		NetworkPolicy: np,
	}
}

// WithPodSelector sets the pod selector for the NetworkPolicy
func (builder *NetworkPolicyBuilder) WithPodSelector(matchLabels map[string]string) *NetworkPolicyBuilder {
	builder.NetworkPolicy.Spec.PodSelector = metav1.LabelSelector{
		MatchLabels: matchLabels,
	}
	return builder
}

// WithIngressPolicyType adds Ingress to the policy types
func (builder *NetworkPolicyBuilder) WithIngressPolicyType() *NetworkPolicyBuilder {
	if builder.NetworkPolicy.Spec.PolicyTypes == nil {
		builder.NetworkPolicy.Spec.PolicyTypes = []networkingv1.PolicyType{}
	}

	// Check if Ingress is already present
	if slices.Contains(builder.NetworkPolicy.Spec.PolicyTypes, networkingv1.PolicyTypeIngress) {
		return builder
	}

	builder.NetworkPolicy.Spec.PolicyTypes = append(builder.NetworkPolicy.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	return builder
}

// WithEgressPolicyType adds Egress to the policy types
func (builder *NetworkPolicyBuilder) WithEgressPolicyType() *NetworkPolicyBuilder {
	if builder.NetworkPolicy.Spec.PolicyTypes == nil {
		builder.NetworkPolicy.Spec.PolicyTypes = []networkingv1.PolicyType{}
	}

	// Check if Egress is already present
	if slices.Contains(builder.NetworkPolicy.Spec.PolicyTypes, networkingv1.PolicyTypeEgress) {
		return builder
	}

	builder.NetworkPolicy.Spec.PolicyTypes = append(builder.NetworkPolicy.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)
	return builder
}

// AllowAllIngress adds a rule that allows all ingress traffic
func (builder *NetworkPolicyBuilder) AllowAllIngress() *NetworkPolicyBuilder {
	builder.WithIngressPolicyType()
	builder.NetworkPolicy.Spec.Ingress = append(builder.NetworkPolicy.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{})
	return builder
}

// AllowAllEgress adds a rule that allows all egress traffic
func (builder *NetworkPolicyBuilder) AllowAllEgress() *NetworkPolicyBuilder {
	builder.WithEgressPolicyType()
	builder.NetworkPolicy.Spec.Egress = append(builder.NetworkPolicy.Spec.Egress, networkingv1.NetworkPolicyEgressRule{})
	return builder
}

// IngressRuleBuilder helps build ingress rules fluently
type IngressRuleBuilder struct {
	rule                 *networkingv1.NetworkPolicyIngressRule
	networkPolicyBuilder *NetworkPolicyBuilder
}

// NewIngressRule starts building a new ingress rule
func (builder *NetworkPolicyBuilder) NewIngressRule() *IngressRuleBuilder {
	builder.WithIngressPolicyType()
	rule := &networkingv1.NetworkPolicyIngressRule{}
	return &IngressRuleBuilder{
		rule:                 rule,
		networkPolicyBuilder: builder,
	}
}

func (irb *IngressRuleBuilder) OnPort(protocol corev1.Protocol, port int32) *IngressRuleBuilder {
	portRule := networkingv1.NetworkPolicyPort{
		Protocol: &protocol,
		Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: port},
	}
	irb.rule.Ports = append(irb.rule.Ports, portRule)
	return irb
}

// OnNamedPort adds a named port to the ingress rule
func (irb *IngressRuleBuilder) OnNamedPort(protocol corev1.Protocol, portName string) *IngressRuleBuilder {
	portRule := networkingv1.NetworkPolicyPort{
		Protocol: &protocol,
		Port:     &intstr.IntOrString{Type: intstr.String, StrVal: portName},
	}
	irb.rule.Ports = append(irb.rule.Ports, portRule)
	return irb
}

// End completes the ingress rule and returns to the NetworkPolicyBuilder
func (irb *IngressRuleBuilder) End() *NetworkPolicyBuilder {
	irb.networkPolicyBuilder.NetworkPolicy.Spec.Ingress = append(irb.networkPolicyBuilder.NetworkPolicy.Spec.Ingress, *irb.rule)

	for _, i := range irb.networkPolicyBuilder.NetworkPolicy.Spec.Ingress {
		sort.Slice(i.Ports, func(x, y int) bool {
			return json.MustMarshal(i.Ports[x]) < json.MustMarshal(i.Ports[y])
		})
	}
	sort.Slice(irb.networkPolicyBuilder.NetworkPolicy.Spec.Ingress, func(i, j int) bool {
		return len(irb.networkPolicyBuilder.NetworkPolicy.Spec.Ingress[i].Ports) < len(irb.networkPolicyBuilder.NetworkPolicy.Spec.Ingress[j].Ports)
	})

	return irb.networkPolicyBuilder
}

// EgressRuleBuilder helps build egress rules fluently
type EgressRuleBuilder struct {
	rule                 *networkingv1.NetworkPolicyEgressRule
	networkPolicyBuilder *NetworkPolicyBuilder
}

// NewEgressRule starts building a new egress rule
func (builder *NetworkPolicyBuilder) NewEgressRule() *EgressRuleBuilder {
	builder.WithEgressPolicyType()
	rule := &networkingv1.NetworkPolicyEgressRule{}
	return &EgressRuleBuilder{
		rule:                 rule,
		networkPolicyBuilder: builder,
	}
}

// OnPort adds a port to the egress rule
func (erb *EgressRuleBuilder) OnPort(protocol corev1.Protocol, port int32) *EgressRuleBuilder {
	portRule := networkingv1.NetworkPolicyPort{
		Protocol: &protocol,
		Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: port},
	}
	erb.rule.Ports = append(erb.rule.Ports, portRule)
	return erb
}

// OnNamedPort adds a named port to the egress rule
func (erb *EgressRuleBuilder) OnNamedPort(protocol corev1.Protocol, portName string) *EgressRuleBuilder {
	portRule := networkingv1.NetworkPolicyPort{
		Protocol: &protocol,
		Port:     &intstr.IntOrString{Type: intstr.String, StrVal: portName},
	}
	erb.rule.Ports = append(erb.rule.Ports, portRule)
	return erb
}

// End completes the egress rule and returns to the NetworkPolicyBuilder
func (erb *EgressRuleBuilder) End() *NetworkPolicyBuilder {
	erb.networkPolicyBuilder.NetworkPolicy.Spec.Egress = append(erb.networkPolicyBuilder.NetworkPolicy.Spec.Egress, *erb.rule)
	for _, i := range erb.networkPolicyBuilder.NetworkPolicy.Spec.Egress {
		sort.Slice(i.Ports, func(x, y int) bool {
			return json.MustMarshal(i.Ports[x]) < json.MustMarshal(i.Ports[y])
		})
	}
	sort.Slice(erb.networkPolicyBuilder.NetworkPolicy.Spec.Egress, func(i, j int) bool {
		return len(erb.networkPolicyBuilder.NetworkPolicy.Spec.Egress[i].Ports) < len(erb.networkPolicyBuilder.NetworkPolicy.Spec.Egress[j].Ports)
	})

	return erb.networkPolicyBuilder
}

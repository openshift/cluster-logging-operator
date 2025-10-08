package network

import (
	"context"
	"fmt"

	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileClusterLogForwarderNetworkPolicy reconciles the NetworkPolicy for the clusterlogforwarder
// It handles both AllowAllIngressEgress and RestrictIngressEgress rule sets, parsing ports from outputs and inputs when needed.
func ReconcileClusterLogForwarderNetworkPolicy(k8Client client.Client, namespace, policyName, instanceName, component string, policyRuleSet obsv1.NetworkPolicyRuleSetType, outputs []obsv1.OutputSpec, inputs []obsv1.InputSpec, ownerRef metav1.OwnerReference, visitor func(o runtime.Object)) error {
	var egressPorts []factory.PortProtocol
	var ingressPorts []int32

	// For RestrictIngressEgress, determine the ports to use based on URLs in outputs and defaults
	if policyRuleSet == obsv1.NetworkPolicyRuleSetTypeRestrictIngressEgress {
		// Parse ports with protocols from outputs
		if len(outputs) > 0 {
			egressPorts = GetOutputPortsWithProtocols(outputs)
		}
		// Parse ports from inputs (receiver inputs use TCP)
		if len(inputs) > 0 {
			ingressPorts = GetInputPorts(inputs)
		}
	}

	desired := factory.NewNetworkPolicyWithProtocolPorts(namespace, policyName, instanceName, component, string(policyRuleSet), egressPorts, ingressPorts, visitor)
	utils.AddOwnerRefToObject(desired, ownerRef)

	return reconcile.NetworkPolicy(k8Client, desired)
}

// ReconcileLogFileMetricsExporterNetworkPolicy reconciles the NetworkPolicy for the logfilemetricexporter
func ReconcileLogFileMetricsExporterNetworkPolicy(k8Client client.Client, namespace, policyName, instanceName, component string, policyRuleSet loggingv1alpha1.NetworkPolicyRuleSetType, ownerRef metav1.OwnerReference, visitor func(o runtime.Object)) error {
	desired := factory.NewNetworkPolicyWithProtocolPorts(namespace, policyName, instanceName, component, string(policyRuleSet), nil, nil, visitor)
	utils.AddOwnerRefToObject(desired, ownerRef)

	return reconcile.NetworkPolicy(k8Client, desired)
}

func RemoveNetworkPolicy(k8Client client.Client, namespace, name string) error {
	np := runtime.NewNetworkPolicy(namespace, name)
	if err := k8Client.Delete(context.TODO(), np); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failure deleting networkPolicy %s/%s: %v", namespace, name, err)
	}
	return nil
}

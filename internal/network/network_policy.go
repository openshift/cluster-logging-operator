package network

import (
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileNetworkPolicy reconciles the NetworkPolicy for the collector
func ReconcileNetworkPolicy(k8Client client.Client, namespace, policyName, instanceName string, ownerRef metav1.OwnerReference, visitor func(o runtime.Object)) error {
	desired := factory.NewNetworkPolicy(namespace, policyName, instanceName, visitor)
	utils.AddOwnerRefToObject(desired, ownerRef)

	return reconcile.NetworkPolicy(k8Client, desired)
}

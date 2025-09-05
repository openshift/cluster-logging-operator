package log7571networkpolicy

import (
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	Log7571NetworkPolicyName = "networkPolicy-Log7571"
)

func Handle(context internalcontext.ForwarderContext, objVisitor func(o runtime.Object)) (err error) {
	ownerRef := utils.AsOwner(context.Forwarder)
	resourceNames := factory.ResourceNames(*context.Forwarder)

	// Reconcile NetworkPolicy for the collector daemonset
	if err := network.ReconcileNetworkPolicy(context.Client, context.Forwarder.Namespace, fmt.Sprintf("%s-%s", constants.CollectorName, resourceNames.CommonName), context.Forwarder.Name, ownerRef, objVisitor); err != nil {
		log.Error(err, "collector.ReconcileNetworkPolicy")
		return err
	}

	return nil
}

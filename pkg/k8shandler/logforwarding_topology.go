package k8shandler

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	topologyapi "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology"
	normalizertopology "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology/normalizer"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

type edgeTopology struct {
	requestCluster *ClusterLoggingRequest
}

func (clusterRequest *ClusterLoggingRequest) ReconcileLogForwardingTopology(proxyConfig *configv1.Proxy) (err error) {
	desiredTopology := topologyapi.LogForwardingEdgeNormalizationTopology //the legacy default
	deployedTopology := ""

	if value, found := utils.GetAnnotation(topologyapi.EnableDevPreviewLogForwardingTopologyAnnotation, clusterRequest.Cluster.ObjectMeta); found && strings.ToLower(value) == "true" {
		desiredTopology, _ = utils.GetAnnotation(topologyapi.LogForwardingTopologyAnnotation, clusterRequest.Cluster.ObjectMeta)
	}
	desired, desiredTopology := newLogForwarderTopology(desiredTopology, clusterRequest)
	if condition := clusterRequest.Cluster.Status.Conditions.GetCondition(topologyapi.LogForwardingTopologyCondition); condition != nil {
		deployedTopology = string(condition.Reason)
	}
	if deployedTopology != desiredTopology {
		deployed, _ := newLogForwarderTopology(deployedTopology, clusterRequest)
		if err = deployed.Undeploy(); err != nil {
			logger.Errorf("There was an error trying to undeploy the the logforwarding topology: %v", err)
		}
	}
	topologyCondition := topologyapi.NewLogForwardingTopologyCondition(desiredTopology)
	clusterRequest.Cluster.Status.Conditions.SetCondition(topologyCondition)
	if err = clusterRequest.UpdateStatus(clusterRequest.Cluster); err != nil {
		logger.Warnf("Unable to update the logforwarding status: %v", err)
	}
	return desired.Reconcile(proxyConfig)
}

func newLogForwarderTopology(topology string, clusterRequest *ClusterLoggingRequest) (topologyapi.LogForwarderTopology, string) {

	switch topology {
	case topologyapi.LogForwardingCentralNormalizationTopology:
		return normalizertopology.CentralNormalizerTopology{
			OwnerRef:                utils.AsOwner(clusterRequest.Cluster),
			APIClient:               clusterRequest,
			ReconcilePriorityClass:  clusterRequest.createOrUpdateCollectionPriorityClass,
			ReconcileServiceAccount: clusterRequest.createOrUpdateCollectorServiceAccount,
			ReconcileConfigMap:      clusterRequest.CreateOrUpdateConfigMap,
			ReconcileSecrets:        clusterRequest.createOrUpdateFluentdSecret,
			RemovePriorityClass:     func() error { return clusterRequest.RemovePriorityClass(clusterLoggingPriorityClassName) },
			RemoveServiceAccount:    func() error { return clusterRequest.RemoveServiceAccount("logcollector") },
			RemoveSecrets:           func() error { return clusterRequest.RemoveSecret(fluentdName) },
		}, topologyapi.LogForwardingCentralNormalizationTopology
	default:
	}
	return edgeTopology{
		clusterRequest,
	}, topologyapi.LogForwardingEdgeNormalizationTopology
}

func (topology edgeTopology) Reconcile(proxyConfig *configv1.Proxy) (err error) {
	return topology.requestCluster.CreateOrUpdateCollection(proxyConfig)
}

func (topology edgeTopology) Undeploy() error {
	return topology.requestCluster.UndeployCollector()
}

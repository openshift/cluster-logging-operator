package k8shandler

import (
	"strings"

	"github.com/ViaQ/logerr/log"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	topologyapi "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology"
	normalizertopology "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology/central"
	edgetopologyapi "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology/edge"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type edgeTopology struct {
	requestCluster *ClusterLoggingRequest
}

func (clusterRequest *ClusterLoggingRequest) ReconcileLogForwardingTopology(proxyConfig *configv1.Proxy) (err error) {
	desiredTopology := topologyapi.LogForwardingEdgeNormalizationTopology //the legacy default
	deployedTopology := ""

	if value, found := clusterRequest.Cluster.Annotations[topologyapi.EnableDevPreviewLogForwardingTopologyAnnotation]; found && strings.ToLower(value) == "true" {
		desiredTopology, _ = clusterRequest.Cluster.Annotations[topologyapi.LogForwardingTopologyAnnotation]
	}
	desired := newLogForwarderTopology(desiredTopology, clusterRequest)
	if condition := clusterRequest.Cluster.Status.Conditions.GetCondition(topologyapi.LogForwardingTopologyCondition); condition != nil {
		deployedTopology = string(condition.Reason)
	}
	if deployedTopology != desired.Name() {
		deployed := newLogForwarderTopology(deployedTopology, clusterRequest)
		if err = deployed.Undeploy(); err != nil {
			log.Error(err, "There was an error trying to undeploy the the logforwarding topology")
		}
	}
	topologyCondition := topologyapi.NewLogForwardingTopologyCondition(desiredTopology)
	clusterRequest.Cluster.Status.Conditions.SetCondition(topologyCondition)
	if err = clusterRequest.UpdateStatus(clusterRequest.Cluster); err != nil {
		log.V(1).Error(err, "Unable to update the logforwarding status")
	}
	return desired.Reconcile(proxyConfig)
}

func newLogForwarderTopology(topology string, clusterRequest *ClusterLoggingRequest) topologyapi.LogForwarderTopology {

	switch topology {
	case topologyapi.LogForwardingEnhancedEdgeNormalizationTopology:
		return edgetopologyapi.EdgeTopologyEnhanced{
			ReconcileCollector: clusterRequest.CreateOrUpdateCollection,
			GenerateCollectorConfig: func() (string, error) {
				return clusterRequest.GenerateCollectorConfig(logging.LogCollectionTypeFluentbit, topologyapi.LogForwardingEnhancedEdgeNormalizationTopology)
			},
			UndeployCollector: clusterRequest.UndeployCollector,
		}
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
			GenerateCollectorConfig: func() (string, error) {
				return clusterRequest.GenerateCollectorConfig(logging.LogCollectionTypeFluentd, topologyapi.LogForwardingCentralNormalizationTopology)
			},
		}
	default:
	}
	return edgeTopology{
		requestCluster: clusterRequest,
	}
}

func (topology edgeTopology) Reconcile(proxyConfig *configv1.Proxy) (err error) {
	return topology.requestCluster.CreateOrUpdateCollection(proxyConfig, topology)
}

func (topology edgeTopology) Undeploy() error {
	return topology.requestCluster.UndeployCollector()
}

func (topology edgeTopology) Name() string {
	return topologyapi.LogForwardingEdgeNormalizationTopology
}
func (topology edgeTopology) ProcessConfigMap(cm *v1.ConfigMap) *v1.ConfigMap {
	return cm
}
func (topology edgeTopology) ProcessPodSpec(podSpec *v1.PodSpec) *v1.PodSpec {
	return podSpec
}
func (topology edgeTopology) ProcessService(service *v1.Service) *v1.Service {
	return service
}
func (topology edgeTopology) ProcessServiceMonitor(sm *monitoringv1.ServiceMonitor) *monitoringv1.ServiceMonitor {
	return sm
}

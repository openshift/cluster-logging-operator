package k8shandler

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	collector "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileCollector(requestCluster *logging.ClusterLogging, collector *collector.CollectorSpec, requestClient client.Client) (err error) {
	logger.Debugf("Reconciling collector: %v", collector)
	clusterRequest := ClusterLoggingRequest{
		client:    requestClient,
		cluster:   requestCluster,
		Collector: collector,
	}
	if err = clusterRequest.createOrUpdateCollectionPriorityClass(); err != nil {
		return
	}

	if _, err = clusterRequest.createOrUpdateCollectorServiceAccount(); err != nil {
		return
	}

	if err = clusterRequest.createOrUpdatePromTailService(); err != nil {
		return
	}

	if err = clusterRequest.createOrUpdatePromTailServiceMonitor(); err != nil {
		return
	}

	if err = clusterRequest.createOrUpdatePromTailConfigMap(); err != nil {
		return
	}

	if err = clusterRequest.createOrUpdatePromTailSecret(); err != nil {
		return
	}

	if err = clusterRequest.createOrUpdatePromTailDaemonset(); err != nil {
		return
	}
	return nil
}

package k8shandler

import (
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	eslogstore "github.com/openshift/cluster-logging-operator/internal/logstore/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateLogStore() (err error) {
	if clusterRequest.Cluster.Spec.LogStore == nil {
		return nil
	}

	switch clusterRequest.Cluster.Spec.LogStore.Type {
	case logging.LogStoreTypeElasticsearch:
		fetchClusterLogging := func() (*logging.ClusterLogging, error) {
			instance, err, _ := loader.FetchClusterLogging(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name, true)
			return &instance, err
		}
		return eslogstore.Reconcile(clusterRequest.Client, clusterRequest.Cluster.Spec.LogStore, clusterRequest.Cluster.Namespace, clusterRequest.ResourceNames.InternalLogStoreSecret, utils.AsOwner(clusterRequest.Cluster), fetchClusterLogging)
	case logging.LogStoreTypeLokiStack:
		if clusterRequest.Cluster.DeletionTimestamp == nil {
			return lokistack.ReconcileLokiWriteRbac(clusterRequest.Client)
		}
	default:
	}

	return nil
}

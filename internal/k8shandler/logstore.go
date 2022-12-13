package k8shandler

import (
	eslogstore "github.com/openshift/cluster-logging-operator/internal/logstore/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	"sync"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateLogStore() (err error) {
	if clusterRequest.Cluster.Spec.LogStore == nil {
		return nil
	}

	switch clusterRequest.Cluster.Spec.LogStore.Type {
	case logging.LogStoreTypeElasticsearch:
		fetchClusterLogging := func() (*logging.ClusterLogging, error) {
			return clusterRequest.getClusterLogging(true)
		}
		return eslogstore.Reconcile(clusterRequest.Client, clusterRequest.Cluster.Spec.LogStore, clusterRequest.Cluster.Namespace, utils.AsOwner(clusterRequest.Cluster), fetchClusterLogging)
	case logging.LogStoreTypeLokiStack:
		return lokistack.ReconcileLokiStackLogStore(clusterRequest.Client, clusterRequest.Cluster.DeletionTimestamp, clusterRequest.appendFinalizer)
	default:
		return nil
	}
}

func LoadElasticsearchSecretMap() map[string][]byte {
	var results = map[string][]byte{}
	_ = Syncronize(func() error {
		results = map[string][]byte{
			"elasticsearch.key": utils.GetWorkingDirFileContents("elasticsearch.key"),
			"elasticsearch.crt": utils.GetWorkingDirFileContents("elasticsearch.crt"),
			"logging-es.key":    utils.GetWorkingDirFileContents("logging-es.key"),
			"logging-es.crt":    utils.GetWorkingDirFileContents("logging-es.crt"),
			"admin-key":         utils.GetWorkingDirFileContents("system.admin.key"),
			"admin-cert":        utils.GetWorkingDirFileContents("system.admin.crt"),
			"admin-ca":          utils.GetWorkingDirFileContents("ca.crt"),
		}
		return nil
	})
	return results
}

var mutex sync.Mutex

//Syncronize blocks single threads access using the certificate mutex
func Syncronize(action func() error) error {
	mutex.Lock()
	defer mutex.Unlock()
	return action()
}

package k8shandler

import (
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
)

func (clusterRequest *ClusterLoggingRequest) getKibanaStatus() ([]elasticsearch.KibanaStatus, error) {
	cr, err := clusterRequest.getKibanaCR()
	if err != nil {
		return nil, err
	}
	return cr.Status, nil
}

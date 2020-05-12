package k8shandler

import (
	"fmt"

	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ElasticsearchRequest struct {
	client          client.Client
	cluster         *elasticsearch.Elasticsearch
	FnCurlEsService func(clusterName, namespace string, payload *esCurlStruct, client client.Client)
}

func Reconcile(requestCluster *elasticsearch.Elasticsearch, requestClient client.Client) error {

	elasticsearchRequest := ElasticsearchRequest{
		client:          requestClient,
		cluster:         requestCluster,
		FnCurlEsService: curlESService,
	}

	// Ensure existence of servicesaccount
	if err := elasticsearchRequest.CreateOrUpdateServiceAccount(); err != nil {
		return fmt.Errorf("Failed to reconcile ServiceAccount for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of clusterroles and clusterrolebindings
	if err := elasticsearchRequest.CreateOrUpdateRBAC(); err != nil {
		return fmt.Errorf("Failed to reconcile Roles and RoleBindings for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of config maps
	if err := elasticsearchRequest.CreateOrUpdateConfigMaps(); err != nil {
		return fmt.Errorf("Failed to reconcile ConfigMaps for Elasticsearch cluster: %v", err)
	}

	if err := elasticsearchRequest.CreateOrUpdateServices(); err != nil {
		return fmt.Errorf("Failed to reconcile Services for Elasticsearch cluster: %v", err)
	}

	// Ensure Elasticsearch cluster itself is up to spec
	//if err = k8shandler.CreateOrUpdateElasticsearchCluster(cluster, "elasticsearch", "elasticsearch"); err != nil {
	if err := elasticsearchRequest.CreateOrUpdateElasticsearchCluster(); err != nil {
		return fmt.Errorf("Failed to reconcile Elasticsearch deployment spec: %v", err)
	}

	// Ensure existence of service monitors
	if err := elasticsearchRequest.CreateOrUpdateServiceMonitors(); err != nil {
		return fmt.Errorf("Failed to reconcile Service Monitors for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of prometheus rules
	if err := elasticsearchRequest.CreateOrUpdatePrometheusRules(); err != nil {
		return fmt.Errorf("Failed to reconcile PrometheusRules for Elasticsearch cluster: %v", err)
	}

	// Ensure index management is in place
	if err := elasticsearchRequest.CreateOrUpdateIndexManagement(); err != nil {
		return fmt.Errorf("Failed to reconcile IndexMangement for Elasticsearch cluster: %v", err)
	}

	return nil
}

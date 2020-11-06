package k8shandler

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/pkg/indexmanagement"
	"github.com/openshift/elasticsearch-operator/pkg/logger"
	esapi "github.com/openshift/elasticsearch-operator/pkg/types/elasticsearch"
)

const (
	//ocpTemplatePrefix is the prefix all operator generated templates
	ocpTemplatePrefix = "ocp-gen"
)

func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdateIndexManagement() error {

	logger.Debug("Reconciling IndexManagement")
	cluster := elasticsearchRequest.cluster
	if cluster.Spec.IndexManagement == nil {
		logger.Debug("IndexManagement not specified - noop")
		return nil
	}
	spec := indexmanagement.VerifyAndNormalize(cluster)
	policies := spec.PolicyMap()
	elasticsearchRequest.cullIndexManagement(spec.Mappings, policies)
	for _, mapping := range spec.Mappings {
		logger.Debugf("reconciling index management for mapping: %s", mapping.Name)
		//create or update template
		if err := elasticsearchRequest.createOrUpdateIndexTemplate(mapping); err != nil {
			logger.Errorf("Error creating index template for mapping %s: %v", mapping.Name, err)
			return err
		}
		//TODO: Can we have partial success?
		if err := elasticsearchRequest.initializeIndexIfNeeded(mapping); err != nil {
			logger.Errorf("Error intializing index for mapping %s: %v", mapping.Name, err)
			return err
		}
	}
	if err := indexmanagement.ReconcileCurationConfigmap(elasticsearchRequest.client, elasticsearchRequest.cluster); err != nil {
		return err
	}
	primaryShards := getDataCount(elasticsearchRequest.cluster)
	for _, mapping := range spec.Mappings {
		policy := policies[mapping.PolicyRef]
		if err := indexmanagement.ReconcileRolloverCronjob(elasticsearchRequest.client, elasticsearchRequest.cluster, policy, mapping, primaryShards); err != nil {
			logger.Errorf("There was an error reconciling the rollover cronjob for policy %q: %v", policy.Name, err)
			return err
		}
		if err := indexmanagement.ReconcileCurationCronjob(elasticsearchRequest.client, elasticsearchRequest.cluster, policy, mapping, primaryShards); err != nil {
			logger.Errorf("There was an error reconciling the curation cronjob for policy %q: %v", policy.Name, err)
			return err
		}
	}

	return nil
}

func (elasticsearchRequest *ElasticsearchRequest) cullIndexManagement(mappings []logging.IndexManagementPolicyMappingSpec, policies logging.PolicyMap) {
	cluster := elasticsearchRequest.cluster
	client := elasticsearchRequest.client
	esClient := elasticsearchRequest.esClient

	if err := indexmanagement.RemoveCronJobsForMappings(client, cluster, mappings, policies); err != nil {
		logger.Errorf("Unable to cull cronjobs: %v", err)
	}
	mappingNames := sets.NewString()
	for _, mapping := range mappings {
		mappingNames.Insert(formatTemplateName(mapping.Name))
	}
	existing, err := esClient.ListTemplates()
	if err != nil {
		logger.Warnf("Unable to list existing templates in order to reconcile stale ones: %v", err)
		return
	}
	difference := existing.Difference(mappingNames)

	for _, template := range difference.List() {
		if strings.HasPrefix(template, ocpTemplatePrefix) {
			if err := esClient.DeleteIndexTemplate(template); err != nil {
				logger.Warnf("Unable to delete stale template %q in order to reconcile: %v", template, err)
			}
		}
	}
}
func (elasticsearchRequest *ElasticsearchRequest) initializeIndexIfNeeded(mapping logging.IndexManagementPolicyMappingSpec) error {
	cluster := elasticsearchRequest.cluster
	esClient := elasticsearchRequest.esClient

	pattern := formatWriteAlias(mapping)
	indices, err := esClient.ListIndicesForAlias(pattern)
	if err != nil {
		return err
	}
	if len(indices) < 1 {
		indexName := fmt.Sprintf("%s-000001", mapping.Name)
		primaryShards := getDataCount(cluster)
		replicas := int32(calculateReplicaCount(cluster))
		index := esapi.NewIndex(indexName, primaryShards, replicas)
		index.AddAlias(mapping.Name, false)
		index.AddAlias(pattern, true)
		for _, alias := range mapping.Aliases {
			index.AddAlias(alias, false)
		}
		return esClient.CreateIndex(indexName, index)
	}
	return nil
}

func formatTemplateName(name string) string {
	return fmt.Sprintf("%s-%s", ocpTemplatePrefix, name)
}

func formatWriteAlias(mapping logging.IndexManagementPolicyMappingSpec) string {
	return fmt.Sprintf("%s-write", mapping.Name)
}

func (elasticsearchRequest *ElasticsearchRequest) createOrUpdateIndexTemplate(mapping logging.IndexManagementPolicyMappingSpec) error {
	cluster := elasticsearchRequest.cluster
	esClient := elasticsearchRequest.esClient

	name := formatTemplateName(mapping.Name)
	pattern := fmt.Sprintf("%s*", mapping.Name)
	primaryShards := getDataCount(cluster)
	replicas := int32(calculateReplicaCount(cluster))
	aliases := append(mapping.Aliases, mapping.Name)
	template := esapi.NewIndexTemplate(pattern, aliases, primaryShards, replicas)

	return esClient.CreateIndexTemplate(name, template)
}

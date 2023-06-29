package elasticsearch

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/logstore/elasticsearch/indexmanagement"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	maximumElasticsearchMasterCount = int32(3)
)

var (
	defaultEsMemory     resource.Quantity = resource.MustParse("16Gi")
	defaultEsCpuRequest resource.Quantity = resource.MustParse("1")

	defaultEsProxyMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultEsProxyCpuRequest resource.Quantity = resource.MustParse("100m")
)

func NewEmptyElasticsearchCR(namespace, name, logstoreSecretName string) *elasticsearch.Elasticsearch {
	return &elasticsearch.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"logging.openshift.io/elasticsearch-cert-management":            "true",
				"logging.openshift.io/elasticsearch-cert." + logstoreSecretName: "system.logging.fluentd",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.GroupVersion.String(),
		},
		Spec: elasticsearch.ElasticsearchSpec{},
	}
}

func NewElasticsearchCR(logStore *logging.LogStoreSpec, namespace, name, logstoreSecretName string, existing *elasticsearch.Elasticsearch, ownerRef metav1.OwnerReference) *elasticsearch.Elasticsearch {

	var esNodes []elasticsearch.ElasticsearchNode
	logStoreSpec := logging.LogStoreSpec{
		Elasticsearch: &logging.ElasticsearchSpec{},
	}
	if logStore != nil {
		logStoreSpec = *logStore
	}
	if logStoreSpec.Elasticsearch == nil {
		logStoreSpec.Elasticsearch = &logging.ElasticsearchSpec{}
	}
	var resources = logStoreSpec.Elasticsearch.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: defaultEsMemory,
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultEsMemory,
				v1.ResourceCPU:    defaultEsCpuRequest,
			},
		}
	}
	var proxyResources = logStoreSpec.Elasticsearch.ProxySpec.Resources
	if proxyResources == nil {
		proxyResources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: defaultEsProxyMemory,
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultEsProxyMemory,
				v1.ResourceCPU:    defaultEsProxyCpuRequest,
			},
		}
	}

	esNode := elasticsearch.ElasticsearchNode{
		Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
		NodeCount: logStoreSpec.Elasticsearch.NodeCount,
		Storage:   logStoreSpec.Elasticsearch.Storage,
	}

	// build Nodes
	esNodes = append(esNodes, esNode)

	// if we had more than 1 es node before, we also want to enter this condition
	if logStoreSpec.Elasticsearch.NodeCount > maximumElasticsearchMasterCount || (existing != nil && len(existing.Spec.Nodes) > 1) {

		// we need to check this because if we scaled down we can enter this block
		if logStoreSpec.Elasticsearch.NodeCount > maximumElasticsearchMasterCount {
			esNodes[0].NodeCount = maximumElasticsearchMasterCount
		}

		remainder := logStoreSpec.Elasticsearch.NodeCount - maximumElasticsearchMasterCount
		if remainder < 0 {
			remainder = 0
		}

		dataNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data"},
			NodeCount: remainder,
			Storage:   logStoreSpec.Elasticsearch.Storage,
		}

		esNodes = append(esNodes, dataNode)

	}

	redundancyPolicy := logStoreSpec.Elasticsearch.RedundancyPolicy
	if redundancyPolicy == "" {
		redundancyPolicy = elasticsearch.ZeroRedundancy
	}

	indexManagementSpec := indexmanagement.NewSpec(logStoreSpec.RetentionPolicy)

	es := NewEmptyElasticsearchCR(namespace, name, logstoreSecretName)
	es.Spec = elasticsearch.ElasticsearchSpec{
		Spec: elasticsearch.ElasticsearchNodeSpec{
			Resources:      *resources,
			NodeSelector:   logStoreSpec.Elasticsearch.NodeSelector,
			Tolerations:    logStoreSpec.Elasticsearch.Tolerations,
			ProxyResources: *proxyResources,
		},
		Nodes:            esNodes,
		ManagementState:  elasticsearch.ManagementStateManaged,
		RedundancyPolicy: redundancyPolicy,
		IndexManagement:  indexManagementSpec,
	}

	utils.AddOwnerRefToObject(es, ownerRef)

	return es
}

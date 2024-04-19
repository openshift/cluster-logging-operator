package elasticsearch

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/framework"
)

const (
	ManagedLogStore       = "elasticsearch-rh"
	ManagedLogStoreSecret = "elasticsearch"
	elasticsearchURI      = "apis/logging.openshift.io/v1/namespaces/%s/elasticsearches/%s"
)

type ManagedElasticsearch struct {
	*ElasticLogStore
}

func NewManagedElasticsearch(test framework.Test) *ManagedElasticsearch {
	return &ManagedElasticsearch{
		ElasticLogStore: NewElasticLogStore(test),
	}
}

// TODO: Update to also support deployment of elasticsearch operator
func (es *ManagedElasticsearch) Deploy() error {
	yaml := `
apiVersion: "logging.openshift.io/v1"
kind: "Elasticsearch"
metadata:
  name: "elasticsearch"
  annotations:
      elasticsearch.openshift.io/loglevel: trace
      logging.openshift.io/elasticsearch-cert-management: "true"
      logging.openshift.io/elasticsearch-cert.mycollector: "system.logging.fluentd"
spec:
  managementState: "Managed"
  nodeSpec:
    resources:
      limits:
        memory: 2Gi
      requests:
        cpu: 100m
        memory: 2Gi
  nodes:
  - nodeCount: 1
    roles:
    - client
    - data
    - master
    storage: {}
  redundancyPolicy: ZeroRedundancy
`
	uri := fmt.Sprintf(elasticsearchURI, constants.OpenshiftNS, "elasticsearch")
	es.AddCleanup(func() error {
		return es.Client().RESTClient().Delete().
			RequestURI(uri).
			SetHeader("Content-Type", "application/yaml").
			Do(context.TODO()).Error()
	})
	return es.Client().RESTClient().Post().
		RequestURI(uri).
		SetHeader("Content-Type", "application/yaml").
		Body([]byte(yaml)).
		Do(context.TODO()).Error()
}

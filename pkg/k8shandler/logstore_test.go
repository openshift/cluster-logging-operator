package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewElasticsearchCRWhenResourcesAreUndefined(t *testing.T) {

	cluster := NewClusterLogging(&logging.ClusterLogging{})
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	//check defaults
	resources := elasticsearchCR.Spec.Spec.Resources
	if resources.Limits[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultEsMemory)
	}
	if resources.Limits[v1.ResourceCPU] != defaultEsCpuRequest {
		t.Errorf("Exp. the default CPU limit to be %v", defaultEsCpuRequest)
	}
	if resources.Requests[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultEsMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultEsCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultEsCpuRequest)
	}
}

func TestNewElasticsearchCRWhenNodeSelectorIsDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo": "bar",
	}
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeSelector: expSelector,
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	for _, node := range elasticsearchCR.Spec.Nodes {
		if !reflect.DeepEqual(node.NodeSelector, expSelector) {
			t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, node.NodeSelector)
		}

	}
}

func TestNewElasticsearchCRWhenResourcesAreDefined(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	//check defaults
	resources := elasticsearchCR.Spec.Spec.Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the default memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceCPU] == defaultEsCpuRequest {
		t.Errorf("Exp. the default CPU limit to not be %v", defaultEsCpuRequest)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the default memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the default CPU request to be %v", requestCPU)
	}
}

func TestDifferenceFoundWhenResourcesAreChanged(t *testing.T) {

	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	cluster = NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						Resources: newResourceRequirements("10Gi", "", "12Gi", "500m"),
					},
				},
			},
		},
	)
	elasticsearchCR2 := cluster.newElasticsearchCR("test-app-name")

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to resource change")
	}
}

func TestDifferenceFoundWhenNodeCountIsChanged(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 1,
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	cluster = NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 2,
					},
				},
			},
		},
	)
	elasticsearchCR2 := cluster.newElasticsearchCR("test-app-name")

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}
}

func TestDefaultRedundancyUsedWhenOmitted(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type:              "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	if !reflect.DeepEqual(elasticsearchCR.Spec.RedundancyPolicy, elasticsearch.ZeroRedundancy) {
		t.Errorf("Exp. the redundancyPolicy to be %q but was %q", elasticsearch.ZeroRedundancy, elasticsearchCR.Spec.RedundancyPolicy)
	}
}

func TestUseRedundancyWhenSpecified(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						RedundancyPolicy: elasticsearch.SingleRedundancy,
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	if !reflect.DeepEqual(elasticsearchCR.Spec.RedundancyPolicy, elasticsearch.SingleRedundancy) {
		t.Errorf("Exp. the redundancyPolicy to be %q but was %q", elasticsearch.SingleRedundancy, elasticsearchCR.Spec.RedundancyPolicy)
	}
}

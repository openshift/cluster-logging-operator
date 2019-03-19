package k8shandler

import (
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewElasticsearchCRWhenResourcesAreUndefined(t *testing.T) {

	cluster := &ClusterLogging{&logging.ClusterLogging{}}
	elasticsearch := cluster.newElasticsearchCR("test-app-name")

	//check defaults
	resources := elasticsearch.Spec.Spec.Resources
	if resources.Limits[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultEsMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultEsMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultEsCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultEsCpuRequest)
	}

	//check node overrides
	resources = elasticsearch.Spec.Nodes[0].Resources
	if resources.Limits[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultEsMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultEsMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultEsCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultEsCpuRequest)
	}
}

func TestNewElasticsearchCRWhenResourcesAreDefined(t *testing.T) {
	cluster := &ClusterLogging{
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
	}
	elasticsearch := cluster.newElasticsearchCR("test-app-name")

	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	//check defaults
	resources := elasticsearch.Spec.Spec.Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the default memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the default memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the default CPU request to be %v", requestCPU)
	}

	//check node overrides
	resources = elasticsearch.Spec.Nodes[0].Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the default memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the default memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the default CPU request to be %v", requestCPU)
	}
}

func TestDifferenceFoundWhenResourcesAreChanged(t *testing.T) {

	cluster := &ClusterLogging{
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
	}
	elasticsearch := cluster.newElasticsearchCR("test-app-name")

	cluster = &ClusterLogging{
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
	}
	elasticsearch2 := cluster.newElasticsearchCR("test-app-name")

	_, different := isElasticsearchCRDifferent(elasticsearch, elasticsearch2)
	if !different {
		t.Errorf("Expected that difference would be found due to resource change")
	}
}

func TestDifferenceFoundWhenNodeCountIsChanged(t *testing.T) {
	cluster := &ClusterLogging{
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
	}
	elasticsearch := cluster.newElasticsearchCR("test-app-name")

	cluster = &ClusterLogging{
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
	}
	elasticsearch2 := cluster.newElasticsearchCR("test-app-name")

	_, different := isElasticsearchCRDifferent(elasticsearch, elasticsearch2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}
}

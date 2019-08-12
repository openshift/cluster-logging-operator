package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1"
	esutils "github.com/openshift/elasticsearch-operator/test/utils"
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

	if !reflect.DeepEqual(elasticsearchCR.Spec.Spec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, elasticsearchCR.Spec.Spec.NodeSelector)
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

func TestNotSplitRolesWhenNodeCountIsLt3(t *testing.T) {
	createAndCheckSingleNodeWithNodeCount(t, 2)
}

func TestNotSplitRolesWhenNodeCountIsEq3(t *testing.T) {
	createAndCheckSingleNodeWithNodeCount(t, 3)
}

func TestSplitRolesWhenNodeCountIsGt3(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 4,
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	// verify that we have two nodes
	if len(elasticsearchCR.Spec.Nodes) != 2 {
		t.Errorf("Exp. the number of ES nodes to be %q but was %q", 2, len(elasticsearchCR.Spec.Nodes))
	}

	clientDataMasterFound := false
	clientDataFound := false

	for _, val := range elasticsearchCR.Spec.Nodes {
		// check that one is client + master (size 3)
		if val.NodeCount == 1 {
			// check that one is client + data (size 4)
			expectedNode := elasticsearch.ElasticsearchNode{
				Roles: []elasticsearch.ElasticsearchNodeRole{"client", "data"},
			}

			if !areNodeRolesSame(expectedNode, val) {
				t.Errorf("Exp. the roles to be %q but was %q", expectedNode.Roles, val.Roles)
			} else {
				clientDataFound = true
			}
		} else {
			if val.NodeCount == 3 {
				expectedNode := elasticsearch.ElasticsearchNode{
					Roles: []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
				}

				if !areNodeRolesSame(expectedNode, val) {
					t.Errorf("Exp. the roles to be %q but was %q", expectedNode.Roles, val.Roles)
				} else {
					clientDataMasterFound = true
				}
			} else {
				t.Errorf("Exp. the NodeCount to be %q or %q but was %q", 3, 1, val.NodeCount)
			}
		}
	}

	if !clientDataMasterFound {
		t.Errorf("Exp. client data master node was not found")
	}

	if !clientDataFound {
		t.Errorf("Exp. client data node was not found")
	}
}

func createAndCheckSingleNodeWithNodeCount(t *testing.T, expectedNodeCount int32) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: expectedNodeCount,
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")

	// verify that we have two nodes
	if len(elasticsearchCR.Spec.Nodes) != 1 {
		t.Errorf("Exp. the number of ES nodes to be %q but was %q", 1, len(elasticsearchCR.Spec.Nodes))
	}

	for _, val := range elasticsearchCR.Spec.Nodes {
		// check that one is client + master (size 3)
		if val.NodeCount == expectedNodeCount {
			// check that one is client + data (size 4)
			expectedNode := elasticsearch.ElasticsearchNode{
				Roles: []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
			}

			if !areNodeRolesSame(expectedNode, val) {
				t.Errorf("Exp. the roles to be %q but was %q", expectedNode.Roles, val.Roles)
			}
		} else {
			t.Errorf("Exp. the NodeCount to be %q but was %q", expectedNodeCount, val.NodeCount)
		}
	}
}

func TestDifferenceFoundWhenNodeCountExceeds3(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 3,
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
						NodeCount: 4,
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

func TestDifferenceFoundWhenNodeCountExceeds4(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 4,
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
						NodeCount: 5,
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

func TestGenUUIDPreservedWhenNodeCountExceeds4(t *testing.T) {
	cluster := NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 3,
					},
				},
			},
		},
	)
	elasticsearchCR := cluster.newElasticsearchCR("test-app-name")
	dataUUID := esutils.GenerateUUID()
	elasticsearchCR.Spec.Nodes[0].GenUUID = &dataUUID

	cluster = NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 4,
					},
				},
			},
		},
	)
	elasticsearchCR2 := cluster.newElasticsearchCR("test-app-name")

	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}

	if diffCR.Spec.Nodes[0].GenUUID == nil || *diffCR.Spec.Nodes[0].GenUUID != dataUUID {
		t.Errorf("Expected that original GenUUID would be preserved as %v but was %v", dataUUID, diffCR.Spec.Nodes[0].GenUUID)
	}
}

func TestGenUUIDPreservedWhenNodeCountChanges(t *testing.T) {
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
	dataUUID := esutils.GenerateUUID()
	elasticsearchCR.Spec.Nodes[0].GenUUID = &dataUUID

	cluster = NewClusterLogging(
		&logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: logging.LogStoreSpec{
					Type: "elasticsearch",
					ElasticsearchSpec: logging.ElasticsearchSpec{
						NodeCount: 3,
					},
				},
			},
		},
	)
	elasticsearchCR2 := cluster.newElasticsearchCR("test-app-name")

	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}

	if diffCR.Spec.Nodes[0].GenUUID == nil || *diffCR.Spec.Nodes[0].GenUUID != dataUUID {
		t.Errorf("Expected that original GenUUID would be preserved as %v but was %v", dataUUID, diffCR.Spec.Nodes[0].GenUUID)
	}
}

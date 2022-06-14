package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler/indexmanagement"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	esutils "github.com/openshift/elasticsearch-operator/test/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewElasticsearchCRWhenResourcesAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	//check defaults
	resources := elasticsearchCR.Spec.Spec.Resources
	if resources.Limits[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultEsMemory)
	}
	if cpu, isPresent := resources.Limits[v1.ResourceCPU]; isPresent {
		t.Errorf("Exp. no default CPU limit, but got %v", cpu.String())
	}
	if resources.Requests[v1.ResourceMemory] != defaultEsMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultEsMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultEsCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultEsCpuRequest)
	}
	proxyResources := elasticsearchCR.Spec.Spec.ProxyResources
	if proxyResources.Limits[v1.ResourceMemory] != defaultEsProxyMemory {
		t.Errorf("Exp. the default proxy memory limit to be %v", defaultEsProxyMemory)
	}
	if cpu, isPresent := proxyResources.Limits[v1.ResourceCPU]; isPresent {
		t.Errorf("Exp. no default proxy CPU limit, but got %v", cpu.String())
	}
	if proxyResources.Requests[v1.ResourceMemory] != defaultEsProxyMemory {
		t.Errorf("Exp. the default proxy memory request to be %v", defaultEsProxyMemory)
	}
	if proxyResources.Requests[v1.ResourceCPU] != defaultEsProxyCpuRequest {
		t.Errorf("Exp. the default proxy CPU request to be %v", defaultEsProxyCpuRequest)
	}
}

func TestNewElasticsearchCRWhenNodeSelectorIsDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo": "bar",
	}
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeSelector: expSelector,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	if !reflect.DeepEqual(elasticsearchCR.Spec.Spec.NodeSelector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, elasticsearchCR.Spec.Spec.NodeSelector)
	}

}

func TestNewElasticsearchCRWhenResourcesAreDefined(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Resources: newResourceRequirements("120Gi", "", "100Gi", "500m"),
					ProxySpec: logging.ProxySpec{
						Resources: newResourceRequirements("512Mi", "", "256Mi", "200m"),
					},
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	limitMemory := resource.MustParse("120Gi")
	requestMemory := resource.MustParse("100Gi")
	requestCPU := resource.MustParse("500m")
	//check defaults for elasticsearch
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
	//check defaults for proxy
	limitMemory = resource.MustParse("512Mi")
	requestMemory = resource.MustParse("256Mi")
	requestCPU = resource.MustParse("200m")
	proxyResources := elasticsearchCR.Spec.Spec.ProxyResources
	if proxyResources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the default proxy memory limit to be %v", limitMemory)
	}
	if proxyResources.Requests[v1.ResourceCPU] == defaultEsCpuRequest {
		t.Errorf("Exp. the default proxy CPU limit to not be %v", defaultEsProxyCpuRequest)
	}
	if proxyResources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the default proxy memory request to be %v", requestMemory)
	}
	if proxyResources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the default proxy CPU request to be %v", requestCPU)
	}
}

func TestDifferenceFoundWhenResourcesAreChanged(t *testing.T) {

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Resources: newResourceRequirements("10Gi", "", "12Gi", "500m"),
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to resource change")
	}
}
func TestDifferenceFoundWhenProxyResourcesAreChanged(t *testing.T) {

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Resources: newResourceRequirements("10Gi", "", "12Gi", "500m"),
					ProxySpec: logging.ProxySpec{
						Resources: newResourceRequirements("256Mi", "", "256Mi", "100m"),
					},
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Resources: newResourceRequirements("10Gi", "", "12Gi", "500m"),
					ProxySpec: logging.ProxySpec{
						Resources: newResourceRequirements("512Mi", "", "256Mi", "100m"),
					},
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to resource change")
	}
}

func TestDifferenceFoundWhenNodeCountIsChanged(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 1,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 2,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}
}

func TestDefaultRedundancyUsedWhenOmitted(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type:              "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	if !reflect.DeepEqual(elasticsearchCR.Spec.RedundancyPolicy, elasticsearch.ZeroRedundancy) {
		t.Errorf("Exp. the redundancyPolicy to be %q but was %q", elasticsearch.ZeroRedundancy, elasticsearchCR.Spec.RedundancyPolicy)
	}
}

func TestUseRedundancyWhenSpecified(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					RedundancyPolicy: elasticsearch.SingleRedundancy,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

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
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 4,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

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
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: expectedNodeCount,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

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
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 3,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 4,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}
}

func TestDifferenceFoundWhenNodeCountExceeds4(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 4,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 5,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}
}

func TestNewESCRNoTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type:              "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)
	tolerations := elasticsearchCR.Spec.Spec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewESCRWithTolerations(t *testing.T) {

	expTolerations := []v1.Toleration{
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Tolerations: expTolerations,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)
	tolerations := elasticsearchCR.Spec.Spec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestGenUUIDPreservedWhenNodeCountExceeds4(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 3,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)
	dataUUID := esutils.GenerateUUID()
	elasticsearchCR.Spec.Nodes[0].GenUUID = &dataUUID

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 4,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}

	if diffCR.Spec.Nodes[0].GenUUID == nil || *diffCR.Spec.Nodes[0].GenUUID != dataUUID {
		t.Errorf("Expected that original GenUUID would be preserved as %v but was %v", dataUUID, diffCR.Spec.Nodes[0].GenUUID)
	}
}

func TestGenUUIDPreservedWhenNodeCountChanges(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 1,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)
	dataUUID := esutils.GenerateUUID()
	elasticsearchCR.Spec.Nodes[0].GenUUID = &dataUUID

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 3,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)

	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}

	if diffCR.Spec.Nodes[0].GenUUID == nil || *diffCR.Spec.Nodes[0].GenUUID != dataUUID {
		t.Errorf("Expected that original GenUUID would be preserved as %v but was %v", dataUUID, diffCR.Spec.Nodes[0].GenUUID)
	}
}

func TestESNodesPreservedWhenCountDecFrom4To2(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 4,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)
	dataUUID := esutils.GenerateUUID()
	elasticsearchCR.Spec.Nodes[0].GenUUID = &dataUUID

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 2,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", elasticsearchCR)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}

	if len(elasticsearchCR2.Spec.Nodes) < len(elasticsearchCR.Spec.Nodes) {
		t.Errorf("Expected that the number of es nodes would be preserved as %d but was %d", len(elasticsearchCR.Spec.Nodes), len(elasticsearchCR2.Spec.Nodes))
	}

	if elasticsearchCR2.Spec.Nodes[0].NodeCount != 2 {
		t.Errorf("Expected that the es node count would be 2 but was %d", elasticsearchCR2.Spec.Nodes[0].NodeCount)
	}

	if elasticsearchCR2.Spec.Nodes[1].NodeCount != 0 {
		t.Errorf("Expected that the es node count would be 0 but was %d", elasticsearchCR2.Spec.Nodes[1].NodeCount)
	}
}

func TestESNodesPreservedWhenCountDecFrom3To0(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 3,
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR := cr.newElasticsearchCR("test-app-name", existing)
	dataUUID := esutils.GenerateUUID()
	elasticsearchCR.Spec.Nodes[0].GenUUID = &dataUUID

	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount: 0,
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", elasticsearchCR)

	_, different := isElasticsearchCRDifferent(elasticsearchCR, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to node count change")
	}

	if len(elasticsearchCR2.Spec.Nodes) < len(elasticsearchCR.Spec.Nodes) {
		t.Errorf("Expected that the number of es nodes would be preserved as %d but was %d", len(elasticsearchCR.Spec.Nodes), len(elasticsearchCR2.Spec.Nodes))
	}

	if elasticsearchCR2.Spec.Nodes[0].NodeCount != 0 {
		t.Errorf("Expected that the es node count would be 0 but was %d", elasticsearchCR2.Spec.Nodes[0].NodeCount)
	}
}

func TestIndexManagementChanges(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit("12h"),
					},
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR1 := cr.newElasticsearchCR("test-app-name", existing)
	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					Audit: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit("12h"),
					},
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)
	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR1, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to retention policy change")
	}

	if diffCR.Spec.IndexManagement.Policies[2].Name != indexmanagement.PolicyNameAudit ||
		diffCR.Spec.IndexManagement.Policies[2].Phases.Delete.MinAge != cluster.Spec.LogStore.RetentionPolicy.Audit.MaxAge {
		t.Errorf("Expected that difference would be found due to retention policy change")
	}
}

func TestIndexManagementNamespacePruning(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit("12h"),
					},
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR1 := cr.newElasticsearchCR("test-app-name", existing)
	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge:                  "12h",
						PruneNamespacesInterval: "15m",
						Namespaces: []elasticsearch.IndexManagementDeleteNamespaceSpec{
							{
								Namespace: "openshift-*",
								MinAge:    "1h",
							},
						},
					},
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)
	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR1, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to retention policy change")
	}

	if diffCR.Spec.IndexManagement.Policies[0].Name != indexmanagement.PolicyNameApp ||
		diffCR.Spec.IndexManagement.Policies[0].Phases.Delete.PruneNamespacesInterval != cluster.Spec.LogStore.RetentionPolicy.App.PruneNamespacesInterval ||
		diffCR.Spec.IndexManagement.Policies[0].Phases.Delete.Namespaces[0].Namespace != cluster.Spec.LogStore.RetentionPolicy.App.Namespaces[0].Namespace ||
		diffCR.Spec.IndexManagement.Policies[0].Phases.Delete.Namespaces[0].MinAge != cluster.Spec.LogStore.RetentionPolicy.App.Namespaces[0].MinAge {
		t.Errorf("Expected that difference would be found due to retention policy change")
	}
}

func TestIndexManagementDeleteByPercentage(t *testing.T) {
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit("12h"),
					},
				},
			},
		},
	}
	cr := &ClusterLoggingRequest{
		Cluster: cluster,
	}
	existing := &elasticsearch.Elasticsearch{}
	elasticsearchCR1 := cr.newElasticsearchCR("test-app-name", existing)
	cluster = &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge:               "12h",
						DiskThresholdPercent: 75,
					},
				},
			},
		},
	}
	cr = &ClusterLoggingRequest{
		Cluster: cluster,
	}
	elasticsearchCR2 := cr.newElasticsearchCR("test-app-name", existing)
	diffCR, different := isElasticsearchCRDifferent(elasticsearchCR1, elasticsearchCR2)
	if !different {
		t.Errorf("Expected that difference would be found due to retention policy change")
	}

	if diffCR.Spec.IndexManagement.Policies[0].Name != indexmanagement.PolicyNameApp ||
		diffCR.Spec.IndexManagement.Policies[0].Phases.Delete.DiskThresholdPercent != cluster.Spec.LogStore.RetentionPolicy.App.DiskThresholdPercent {
		t.Errorf("Expected that difference would be found due to retention policy change")
	}
}

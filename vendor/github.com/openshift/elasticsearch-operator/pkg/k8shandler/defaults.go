package k8shandler

import (
	"fmt"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

const (
	modeUnique    = "unique"
	modeSharedOps = "shared_ops"

	defaultMode = modeSharedOps
	// ES
	defaultESCpuRequest    = "100m"
	defaultESMemoryLimit   = "4Gi"
	defaultESMemoryRequest = "1Gi"
	// ESProxy
	defaultESProxyCpuRequest    = "100m"
	defaultESProxyMemoryLimit   = "64Mi"
	defaultESProxyMemoryRequest = "64Mi"

	maxMasterCount = 3

	elasticsearchCertsPath  = "/etc/openshift/elasticsearch/secret"
	elasticsearchConfigPath = "/usr/share/java/elasticsearch/config"
	heapDumpLocation        = "/elasticsearch/persistent/heapdump.hprof"

	k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	yellowClusterState = "yellow"
	greenClusterState  = "green"
)

var desiredClusterStates = []string{yellowClusterState, greenClusterState}

func kibanaIndexMode(mode string) (string, error) {
	if mode == "" {
		return defaultMode, nil
	}
	if mode == modeUnique || mode == modeSharedOps {
		return mode, nil
	}
	return "", fmt.Errorf("invalid kibana index mode provided [%s]", mode)
}

func esUnicastHost(clusterName, namespace string) string {
	return fmt.Sprintf("%v-cluster.%v.svc", clusterName, namespace)
}

func calculateReplicaCount(dpl *api.Elasticsearch) int {
	dataNodeCount := int((getDataCount(dpl)))
	repType := dpl.Spec.RedundancyPolicy
	switch repType {
	case api.FullRedundancy:
		return dataNodeCount - 1
	case api.MultipleRedundancy:
		return (dataNodeCount - 1) / 2
	case api.SingleRedundancy:
		return 1
	case api.ZeroRedundancy:
		return 0
	default:
		if dataNodeCount == 1 {
			return 0
		}
		return 1
	}
}

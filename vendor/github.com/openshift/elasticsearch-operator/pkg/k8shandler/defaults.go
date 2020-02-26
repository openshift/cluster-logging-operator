package k8shandler

import (
	"fmt"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

const (
	modeUnique    = "unique"
	modeSharedOps = "shared_ops"

	defaultMode               = modeSharedOps
	defaultMasterCPURequest   = "100m"
	defaultCPURequest         = "100m"
	defaultMemoryLimit        = "4Gi"
	defaultMemoryRequest      = "1Gi"
	elasticsearchDefaultImage = "quay.io/openshift/origin-logging-elasticsearch6"

	maxMasterCount = 3

	elasticsearchCertsPath  = "/etc/openshift/elasticsearch/secret"
	elasticsearchConfigPath = "/usr/share/java/elasticsearch/config"
	heapDumpLocation        = "/elasticsearch/persistent/heapdump.hprof"

	k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	logAppenderAnnotation = "elasticsearch.openshift.io/develLogAppender"
)

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

func rootLogger(cluster *api.Elasticsearch) string {
	if value, ok := cluster.GetAnnotations()[log4jConfig]; ok {
		return value
	}
	return "rolling"
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

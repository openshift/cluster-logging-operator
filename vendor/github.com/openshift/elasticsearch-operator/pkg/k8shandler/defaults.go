package k8shandler

import (
	"fmt"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
)

const (
	modeUnique    = "unique"
	modeOpsShared = "ops_shared"
	defaultMode   = modeUnique
)

func kibanaIndexMode(mode string) (string, error) {
	if mode == "" {
		return defaultMode, nil
	}
	if mode == modeUnique || mode == modeOpsShared {
		return mode, nil
	}
	return "", fmt.Errorf("invalid kibana index mode provided [%s]", mode)
}

func esUnicastHost(clusterName string) string {
	return clusterName + "-cluster"
}

func rootLogger() string {
	return "rolling"
}

func calculateReplicaCount(dpl *v1alpha1.Elasticsearch) int {
	dataNodeCount := int((getDataCount(dpl)))
	repType := dpl.Spec.ReplicationPolicy
	switch repType {
	case v1alpha1.FullReplication:
		return dataNodeCount - 1
	case v1alpha1.PartialReplication:
		return (dataNodeCount - 1) / 2
	case v1alpha1.NoReplication:
		fallthrough
	default:
		return 0
	}
}

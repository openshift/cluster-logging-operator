package k8shandler

import (
	"fmt"
)

const (
	modeUnique    = "unique"
	modeOpsShared = "ops_shared"
	defaultMode   = modeUnique
)

func KibanaIndexMode(mode string) (string, error) {
	if mode == "" {
		return defaultMode, nil
	}
	if mode == modeUnique || mode == modeOpsShared {
		return mode, nil
	}
	return "", fmt.Errorf("invalid kibana index mode provided [%s]", mode)
}

func EsUnicastHost(clusterName string) string {
	return clusterName + "-cluster"
}

func RootLogger() string {
	return "rolling"
}

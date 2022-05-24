//go:build fluentd
// +build fluentd

package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func init() {
	LogCollectionType = logging.LogCollectionTypeFluentd
}

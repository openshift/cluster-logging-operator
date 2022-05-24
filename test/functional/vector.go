//go:build vector
// +build vector

package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func init() {
	LogCollectionType = logging.LogCollectionTypeVector
}

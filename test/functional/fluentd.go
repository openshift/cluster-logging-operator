//go:build !vector
// +build !vector

// NOTE: fluentd is currently the default if no build tag (vector or fluetnd) is specified.
// This will switch to vector eventually.

package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func init() {
	LogCollectionType = logging.LogCollectionTypeFluentd
}

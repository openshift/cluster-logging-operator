//go:build fluentd

package functional

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

const LogCollectionType = logging.LogCollectionTypeFluentd

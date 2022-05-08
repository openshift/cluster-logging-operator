package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/fluentd"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/vector"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

func NewApplicationLogTemplate(collector logging.LogCollectionType) types.ApplicationLog {
	if collector == logging.LogCollectionTypeFluentd {
		return fluentd.NewApplicationLogTemplate()
	}
	if collector == logging.LogCollectionTypeVector {
		return vector.NewApplicationLogTemplate()
	}
	return types.ApplicationLog{}
}

func NewContainerInfrastructureLogTemplate(collector logging.LogCollectionType) types.ApplicationLog {
	if collector == logging.LogCollectionTypeFluentd {
		return fluentd.NewContainerInfrastructureLogTemplate()
	}
	if collector == logging.LogCollectionTypeVector {
		return vector.NewContainerInfrastructureLogTemplate()
	}
	return types.ApplicationLog{}
}

package input

import (
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input/container"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func New(input *adapters.Input, resNames factory.ForwarderResourceNames, secrets internalobs.Secrets, op utils.Options) (receivers api.Receivers) {
	framework.SetTLSProfileOptionsFrom(op, input)
	receivers = api.Receivers{}
	switch {
	case input.IsContainerSource():
		receivers.Add(container.New(input.InputSpec))
	}

	return receivers
}

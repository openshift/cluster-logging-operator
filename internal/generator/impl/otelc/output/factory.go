package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/output/lokistack"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/output/otlp"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

func New(o *adapters.Output, inputs []obs.InputSpec, secrets map[string]*corev1.Secret, op utils.Options) (api.Exporters, api.Extensions) {

	exporters := api.Exporters{}
	extensions := api.Extensions{}

	switch o.Type {
	case obs.OutputTypeLokiStack:
		e, x := lokistack.New(o.Name, o.OutputSpec, inputs, secrets, op)
		exporters.Merge(e)
		extensions.Merge(x)
	case obs.OutputTypeOTLP:
		e, x := otlp.New(o, secrets)
		exporters.Merge(e)
		extensions.Merge(x)
	}
	return exporters, extensions
}

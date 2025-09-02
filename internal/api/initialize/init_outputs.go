package initialize

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
)

// MigrateOutputs initializes the outputs for a ClusterLogForwarder
func MigrateOutputs(forwarder obs.ClusterLogForwarder, options utils.Options) obs.ClusterLogForwarder {
	enabled := common.IsEnabledAnnotation(forwarder, constants.AnnotationOtlpOutputTechPreview)
	for _, o := range forwarder.Spec.Outputs {
		if enabled {
			if o.Type == obs.OutputTypeLokiStack && o.LokiStack != nil && o.LokiStack.DataModel == "" {
				o.LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
			}
		} else if o.Type == obs.OutputTypeLokiStack && o.LokiStack != nil {
			o.LokiStack.DataModel = obs.LokiStackDataModelViaq
		}
	}
	return forwarder
}

package clusterlogging

import (
	v1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateClusterLoggingSpec(cl v1.ClusterLogging, k8sclient client.Client) error {

	if cl.Namespace != constants.OpenshiftNS && cl.Name != constants.SingletonName {
		return errors.NewValidationError("Only ClusterLogging named %s/%s is supported", cl.Namespace, cl.Name)
	}
	spec := cl.Spec
	if spec.Forwarder != nil || spec.Curation != nil || spec.Collection != nil {
		return errors.NewValidationError("Only spec.logsStore.type loki with spec.visualization.type ocpConsole is supported for deploying visualization")
	}
	if spec.LogStore != nil && spec.LogStore.Type != v1.LogStoreTypeLokiStack {
		return errors.NewValidationError("spec.logsStore.type loki must be defined to deploy visualization")
	}
	if spec.Visualization != nil && spec.Visualization.Type != v1.VisualizationTypeOCPConsole {
		return errors.NewValidationError("only spec.visualization.type ocpConsole is supported to deploy visualization")
	}
	return nil
}

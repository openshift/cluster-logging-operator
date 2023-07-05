package clusterlogging

import (
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
)

func validateClusterLoggingSpec(cl v1.ClusterLogging) error {

	if cl.Namespace == constants.OpenshiftNS && cl.Name == constants.SingletonName {
		return nil
	}
	spec := cl.Spec
	if spec.Forwarder != nil || spec.LogStore != nil || spec.Curation != nil || spec.Visualization != nil {
		return errors.NewValidationError("Only spec.collection is allowed when using multiple instances of ClusterLogForwarder: %s/%s", cl.Namespace, cl.Name)
	}
	if spec.Collection.Logs != nil {
		return errors.NewValidationError("The use of spec.collection.logs is deprecated in favor of spec.collection fields of ClusterLogForwarder: %s/%s", cl.Namespace, cl.Name)
	}
	if spec.Collection.Type != v1.LogCollectionTypeVector {
		return errors.NewValidationError("Only vector collector impl is supported when using multiple instances of ClusterLogForwarder: %s/%s", cl.Namespace, cl.Name)
	}
	return nil
}

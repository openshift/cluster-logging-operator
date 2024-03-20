package clusterlogging

import (
	"context"
	"fmt"

	v1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateClusterLoggingSpec(cl v1.ClusterLogging, k8sClient client.Client) error {

	if cl.Namespace == constants.OpenshiftNS && cl.Name == constants.SingletonName {
		if cl.Spec.Collection != nil && !cl.Spec.Collection.Type.IsSupportedCollector() {
			return errors.NewValidationError("Collector implementation is not supported: %q", cl.Spec.Collection.Type)
		}
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

//CL(openshift-logging/instance) only - This is a valid LEGACY use case
//CL(openshift-logging/instance) & CLF(openshift-logging/instance) - This is a valid LEGACY usecase
//CL(ANY_NS/OTHER) - This is an invalid use case
//CLF(ANY_NS/ANY_NAME) - This is a valid mCLF use case
//CL(ANY_NS/ANY_NAME) && CLF(ANY_NS/ANY_NAME) - This is a valid mCLF use case
func validateSetup(cl v1.ClusterLogging, k8sClient client.Client) error {
	key := types.NamespacedName{Name: cl.Name, Namespace: cl.Namespace}
	clf := &v1.ClusterLogForwarder{}

	// Check if ClusterLogForwarder exists
	err := k8sClient.Get(context.TODO(), key, clf)
	isCLFNotFound := apierrors.IsNotFound(err)
	if isCLFNotFound {
		// Determine if it's a legacy or mCLF instance
		isLegacy := cl.Namespace == constants.OpenshiftNS && cl.Name == constants.SingletonName

		if isLegacy {
			return nil // Legacy ClusterLogging with no ClusterLogForwarder, valid use case
		} else {
			msg := fmt.Sprintf("ClusterLogging instance requires to have a ClusterLogForwarder deployed in the same namespace and named the same, if not eqauls %s/%s", constants.OpenshiftNS, constants.SingletonName)
			return errors.NewValidationError(msg)
		}
	}

	return nil // Legacy ClusterLogging with existing ClusterLogForwarder or mCLF ClusterLogForwarder, valid use case
}

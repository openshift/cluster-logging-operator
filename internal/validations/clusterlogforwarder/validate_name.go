package clusterlogforwarder

import (
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateName(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus) {

	if clf.Namespace == constants.OpenshiftNS && clf.Name == constants.CollectorName {
		return errors.NewValidationError("Name %q conflicts with an object for the legacy ClusterLogForwarder deployment.  Choose another", clf.Name), nil
	}

	if nameErrors := validation.IsDNS1035Label(clf.Name); len(nameErrors) != 0 {
		return errors.NewValidationError("Name %q will result in an invalid object: %v", clf.Name, nameErrors), nil
	}

	return nil, nil
}

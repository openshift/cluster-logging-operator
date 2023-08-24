package clusterlogforwarder

import (
	"context"
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	//allNamesapces is used for determining cluster scoped bindings
	allNamespaces = ""
)

func ValidateServiceAccount(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	// Do not need to validate SA if legacy forwarder
	if clf.Name == constants.SingletonName && clf.Namespace == constants.OpenshiftNS {
		log.V(3).Info("[ValidateServiceAccount] do not need to validate SA for legacy CL & CLF")
		return nil, nil
	}

	if clf.Namespace == constants.OpenshiftNS && clf.Spec.ServiceAccountName == constants.CollectorServiceAccountName {
		return errors.NewValidationError(constants.CollectorServiceAccountName + " is a reserved serviceaccount name for legacy ClusterLogForwarder(openshift-logging/instance)"), nil
	}

	if clf.Spec.ServiceAccountName == "" {
		return errors.NewValidationError("custom clusterlogforwarders must specify a service account name"), nil
	}

	var err error
	var serviceAccount *corev1.ServiceAccount
	if serviceAccount, err = getServiceAccount(clf.Spec.ServiceAccountName, clf.Namespace, k8sClient); err != nil {
		return err, nil
	}
	// If SA present, validate permissions based off spec'd CLF inputs
	clfInputs := gatherPipelineInputs(clf)
	if err = validateServiceAccountPermissions(k8sClient, clfInputs, serviceAccount, clf.Namespace, clf.Name); err != nil {
		return err, nil
	}
	return nil, nil
}

func getServiceAccount(name, namespace string, k8sClient client.Client) (*corev1.ServiceAccount, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewServiceAccount(namespace, name)
	// Check if service account specified exists
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.NewValidationError("service account not found: %s/%s", namespace, name)
	}
	return proto, nil
}

// ValidateServiceAccountPermissions validates a service account for permissions to collect
// inputs specified by the CLF.
// ie. collect-application-logs, collect-audit-logs, collect-infrastructure-logs
func validateServiceAccountPermissions(k8sClient client.Client, inputs sets.String, serviceAccount *corev1.ServiceAccount, clfNamespace, name string) error {
	if inputs.Len() == 0 {
		err := errors.NewValidationError("There is an error in the input permission validation; no inputs were found to evaluate")
		log.Error(err, "Error while evaluating ClusterLogForwarder permissions", "namespace", clfNamespace, "name", name)
		return err
	}
	var err error
	var username = fmt.Sprintf("system:serviceaccount:%s:%s", serviceAccount.Namespace, serviceAccount.Name)

	// Perform subject access reviews for each spec'd input
	var failedInputs []string
	for _, input := range inputs.List() {
		log.V(3).Info(fmt.Sprintf("[ValidateServiceAccountPermissions] validating %q for user: %v", inputs, username))
		sar := createSubjectAccessReview(username, allNamespaces, "collect", "logs", input, loggingv1.GroupVersion.Group)
		if err = k8sClient.Create(context.TODO(), sar); err != nil {
			return err
		}
		// If input is spec'd but SA isn't authorized to collect it, fail validation
		if !sar.Status.Allowed {
			log.V(3).Info(fmt.Sprintf("[ValidateServiceAccountPermissions] %s %s-logs", errors.NotAuthorizedToCollect, input))
			failedInputs = append(failedInputs, input)
		}
	}

	if len(failedInputs) > 0 {
		return errors.NewValidationError("insufficient permissions on service account, not authorized to collect %q logs", failedInputs)
	}

	return nil
}

func gatherPipelineInputs(clf loggingv1.ClusterLogForwarder) sets.String {
	inputRefs := sets.NewString()
	inputTypes := sets.NewString()

	// Collect inputs from clf pipelines
	for _, pipeline := range clf.Spec.Pipelines {
		for _, input := range pipeline.InputRefs {
			inputRefs.Insert(input)
			if loggingv1.ReservedInputNames.Has(input) {
				inputTypes.Insert(input)
			}
		}
	}

	for _, input := range clf.Spec.Inputs {
		if inputRefs.Has(input.Name) && input.Application != nil {
			inputTypes.Insert(loggingv1.InputNameApplication)
		}
	}

	return *inputTypes
}

func createSubjectAccessReview(user, namespace, verb, resource, name, resourceAPIGroup string) *authorizationapi.SubjectAccessReview {
	sar := &authorizationapi.SubjectAccessReview{
		Spec: authorizationapi.SubjectAccessReviewSpec{
			User: user,
		},
	}
	if strings.HasPrefix(resource, "/") {
		sar.Spec.NonResourceAttributes = &authorizationapi.NonResourceAttributes{
			Path: resource,
			Verb: verb,
		}
	} else {
		sar.Spec.ResourceAttributes = &authorizationapi.ResourceAttributes{
			Resource:  resource,
			Namespace: namespace,
			Group:     resourceAPIGroup,
			Verb:      verb,
			Name:      name,
		}
	}
	return sar
}

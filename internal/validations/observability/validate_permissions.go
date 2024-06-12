package observability

import (
	"context"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
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

var infraNamespaces = regexp.MustCompile(`^default|openshift.*|kube.*$`)

// ValidatePermissions validates the serviceaccount for the CLF has the needed permissions to collect the desired inputs
func ValidatePermissions(context internalcontext.ForwarderContext) (common.AttributeConditionType, []metav1.Condition) {
	clf := context.Forwarder
	k8sClient := context.Client

	var err error
	var serviceAccount *corev1.ServiceAccount
	if serviceAccount, err = getServiceAccount(clf.Spec.ServiceAccount.Name, clf.Namespace, k8sClient); err != nil {
		return common.AttributeConditionConditions, []metav1.Condition{
			internalobs.NewCondition(obs.ConditionAuthorized, obs.ConditionFalse, obs.ReasonServiceAccountDoesNotExist, err.Error()),
		}
	}
	// If SA present, validate permissions based off spec'd CLF inputs
	clfInputs, hasReceiverInputs := gatherPipelineInputs(*clf)
	if err = validateServiceAccountPermissions(k8sClient, clfInputs, hasReceiverInputs, serviceAccount, clf.Namespace, clf.Name); err != nil {
		return common.AttributeConditionConditions, []metav1.Condition{
			internalobs.NewCondition(obs.ConditionAuthorized, obs.ConditionFalse, obs.ReasonClusterRoleMissing, err.Error()),
		}
	}
	return common.AttributeConditionConditions, []metav1.Condition{
		internalobs.NewCondition(obs.ConditionAuthorized, obs.ConditionTrue, obs.ReasonClusterRolesExist, ""),
	}
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
func validateServiceAccountPermissions(k8sClient client.Client, inputs sets.String, hasReceiverInputs bool, serviceAccount *corev1.ServiceAccount, clfNamespace, name string) error {
	if inputs.Len() == 0 && hasReceiverInputs {
		return nil
	}
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
		sar := createSubjectAccessReview(username, allNamespaces, "collect", "logs", input, obs.GroupName)
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

func gatherPipelineInputs(clf obs.ClusterLogForwarder) (sets.String, bool) {
	inputRefs := sets.NewString()
	inputTypes := sets.NewString()

	// Collect inputs from clf pipelines
	for _, pipeline := range clf.Spec.Pipelines {
		for _, input := range pipeline.InputRefs {
			inputRefs.Insert(input)
			if obs.ReservedInputTypes.Has(input) {
				inputTypes.Insert(input)
			}
		}
	}

	noOfReceivers := 0
	for _, input := range clf.Spec.Inputs {

		if inputRefs.Has(input.Name) {
			switch input.Type {
			case obs.InputTypeApplication:
				inputTypes.Insert(string(obs.InputTypeApplication))
				// Check if infra namespaces are spec'd
				if len(input.Application.Includes) > 0 {
					for _, in := range input.Application.Includes {
						if infraNamespaces.MatchString(in.Namespace) {
							inputTypes.Insert(string(obs.InputTypeInfrastructure))
						}
					}
				}
			case obs.InputTypeInfrastructure:
				inputTypes.Insert(string(obs.InputTypeInfrastructure))
			case obs.InputTypeAudit:
				inputTypes.Insert(string(obs.InputTypeAudit))
			case obs.InputTypeReceiver:
				noOfReceivers += 1
				if input.Receiver.Type == obs.ReceiverTypeSyslog {
					inputTypes.Insert(string(obs.InputTypeInfrastructure))
				}
			}
		}
	}

	return *inputTypes, noOfReceivers > 0
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

func hasInfraNamespaces(appNamespaces []string) bool {
	if len(appNamespaces) == 0 {
		return false
	}

	for _, ns := range appNamespaces {
		if infraNamespaces.MatchString(ns) {
			return true
		}
	}
	return false
}

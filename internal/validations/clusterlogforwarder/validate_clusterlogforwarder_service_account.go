package clusterlogforwarder

import (
	"context"
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateServiceAccount(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	// Do not need to validate SA if legacy forwarder
	if clf.Name == constants.SingletonName && clf.Namespace == constants.OpenshiftNS {
		log.V(3).Info("[ValidateServiceAccount] do not need to validate SA for legacy CL & CLF")
		return nil, nil
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
	if err = validateServiceAccountPermissions(k8sClient, clfInputs, serviceAccount, clf.Namespace); err != nil {
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
		return nil, errors.NewValidationError("service account: %s not found", fmt.Sprintf("%q/%q", namespace, name))
	}
	return proto, nil
}

// ValidateServiceAccountPermissions validates a service account for permissions to collect
// inputs specified by the CLF.
// ie. collect-application-logs, collect-audit-logs, collect-infrastructure-logs
func validateServiceAccountPermissions(k8sClient client.Client, inputs sets.String, serviceAccount *corev1.ServiceAccount, clfNamespace string) error {
	var err error
	var username = fmt.Sprintf("system:serviceaccount:%s:%s", serviceAccount.Namespace, serviceAccount.Name)

	// Perform subject access reviews for each spec'd input
	var failedInputs []string
	for input := range inputs {
		log.V(3).Info(fmt.Sprintf("[ValidateServiceAccountPermissions] validating %q for user: %v", inputs, username))
		sar := createSubjectAccessReview(username, clfNamespace, "collect", "logs", input, loggingv1.GroupVersion.Group)
		if err = k8sClient.Create(context.TODO(), sar); err != nil {
			return err
		}
		// If input is spec'd but SA isn't authorized to collect it, fail validation
		if !sar.Status.Allowed {
			log.V(3).Info(fmt.Sprintf("[ValidateServiceAccountPermissions] Not authorized to collect %s-logs", input))
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

	// Collect inputs from clf pipelines
	for _, pipeline := range clf.Spec.Pipelines {
		for _, input := range pipeline.InputRefs {
			inputRefs.Insert(input)
		}
	}
	return inputRefs
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

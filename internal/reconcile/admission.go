package reconcile

import (
	"context"
	"errors"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ValidatingAdmissionPolicy creates or updates a ValidatingAdmissionPolicy.
func ValidatingAdmissionPolicy(ctx context.Context, k8sClient client.Client, desired *admissionregistrationv1.ValidatingAdmissionPolicy) error {
	current := &admissionregistrationv1.ValidatingAdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: desired.Name,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, k8sClient, current, func() error {
		current.Spec = desired.Spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("reconcile ValidatingAdmissionPolicy %q: %w", desired.Name, err)
	}

	log.V(3).Info("reconciled ValidatingAdmissionPolicy", "name", desired.Name, "operation", op)
	return nil
}

// ValidatingAdmissionPolicyBinding creates or updates a ValidatingAdmissionPolicyBinding.
func ValidatingAdmissionPolicyBinding(ctx context.Context, k8sClient client.Client, desired *admissionregistrationv1.ValidatingAdmissionPolicyBinding) error {
	current := &admissionregistrationv1.ValidatingAdmissionPolicyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: desired.Name,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, k8sClient, current, func() error {
		current.Spec = desired.Spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("reconcile ValidatingAdmissionPolicyBinding %q: %w", desired.Name, err)
	}

	log.V(3).Info("reconciled ValidatingAdmissionPolicyBinding", "name", desired.Name, "operation", op)
	return nil
}

// IsUnsupportedAdmissionPolicyAPI returns true when the error indicates the
// ValidatingAdmissionPolicy API is not available on the cluster.
func IsUnsupportedAdmissionPolicyAPI(err error) bool {
	if err == nil {
		return false
	}
	if meta.IsNoMatchError(err) {
		return true
	}
	if apiruntime.IsNotRegisteredError(err) {
		return true
	}
	var groupDiscoveryErr *discovery.ErrGroupDiscoveryFailed
	return errors.As(err, &groupDiscoveryErr)
}

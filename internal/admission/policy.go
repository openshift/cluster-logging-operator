package admission

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

// admissionReconcileBackoff is the shared retry schedule used by the admission
// policy runnables to reconcile the operator's ValidatingAdmissionPolicies once
// the manager cache has started.
var admissionReconcileBackoff = wait.Backoff{
	Steps:    5,
	Duration: 2 * time.Second,
	Factor:   2.0,
	Cap:      30 * time.Second,
}

func decodeValidatingAdmissionPolicy(raw []byte) (*admissionregistrationv1.ValidatingAdmissionPolicy, error) {
	policy := &admissionregistrationv1.ValidatingAdmissionPolicy{}
	if err := yaml.Unmarshal(raw, policy); err != nil {
		return nil, fmt.Errorf("decode ValidatingAdmissionPolicy: %w", err)
	}
	return policy, nil
}

func decodeValidatingAdmissionPolicyBinding(raw []byte) (*admissionregistrationv1.ValidatingAdmissionPolicyBinding, error) {
	binding := &admissionregistrationv1.ValidatingAdmissionPolicyBinding{}
	if err := yaml.Unmarshal(raw, binding); err != nil {
		return nil, fmt.Errorf("decode ValidatingAdmissionPolicyBinding: %w", err)
	}
	return binding, nil
}

func reconcileValidatingAdmissionPolicy(ctx context.Context, k8sClient client.Client, desired *admissionregistrationv1.ValidatingAdmissionPolicy) error {
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
		if isUnsupportedAdmissionPolicyAPI(err) {
			log.Info("ValidatingAdmissionPolicy API is unavailable; skipping admission policy", "name", desired.Name)
			return nil
		}
		return fmt.Errorf("reconcile ValidatingAdmissionPolicy %q: %w", desired.Name, err)
	}

	log.V(3).Info("reconciled ValidatingAdmissionPolicy", "name", desired.Name, "operation", op)
	return nil
}

func reconcileValidatingAdmissionPolicyBinding(ctx context.Context, k8sClient client.Client, desired *admissionregistrationv1.ValidatingAdmissionPolicyBinding) error {
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
		if isUnsupportedAdmissionPolicyAPI(err) {
			log.Info("ValidatingAdmissionPolicyBinding API is unavailable; skipping admission policy binding", "name", desired.Name)
			return nil
		}
		return fmt.Errorf("reconcile ValidatingAdmissionPolicyBinding %q: %w", desired.Name, err)
	}

	log.V(3).Info("reconciled ValidatingAdmissionPolicyBinding", "name", desired.Name, "operation", op)
	return nil
}

func isUnsupportedAdmissionPolicyAPI(err error) bool {
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

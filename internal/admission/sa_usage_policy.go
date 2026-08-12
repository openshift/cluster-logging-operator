package admission

import (
	"context"
	_ "embed"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

const (
	SAUsagePolicyName        = "clf-sa-usage-authorization"
	SAUsagePolicyBindingName = "clf-sa-usage-authorization-binding"
)

//go:embed manifests/clf-sa-usage-authorization.yaml
var saUsagePolicyYAML []byte

//go:embed manifests/clf-sa-usage-authorization-binding.yaml
var saUsagePolicyBindingYAML []byte

// ReconcileSAUsageAuthorization ensures the cluster-scoped ValidatingAdmissionPolicy
// that authorizes ServiceAccount usage in ClusterLogForwarder resources exists.
func ReconcileSAUsageAuthorization(ctx context.Context, k8sClient client.Client) error {
	policy, err := desiredSAUsagePolicy()
	if err != nil {
		return err
	}
	if err := reconcileValidatingAdmissionPolicy(ctx, k8sClient, policy); err != nil {
		return err
	}

	binding, err := desiredSAUsagePolicyBinding()
	if err != nil {
		return err
	}
	return reconcileValidatingAdmissionPolicyBinding(ctx, k8sClient, binding)
}

func desiredSAUsagePolicy() (*admissionregistrationv1.ValidatingAdmissionPolicy, error) {
	return decodeValidatingAdmissionPolicy(saUsagePolicyYAML)
}

func desiredSAUsagePolicyBinding() (*admissionregistrationv1.ValidatingAdmissionPolicyBinding, error) {
	return decodeValidatingAdmissionPolicyBinding(saUsagePolicyBindingYAML)
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
			log.Info("ValidatingAdmissionPolicy API is unavailable; skipping SA usage authorization policy")
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
			log.Info("ValidatingAdmissionPolicyBinding API is unavailable; skipping SA usage authorization binding")
			return nil
		}
		return fmt.Errorf("reconcile ValidatingAdmissionPolicyBinding %q: %w", desired.Name, err)
	}

	log.V(3).Info("reconciled ValidatingAdmissionPolicyBinding", "name", desired.Name, "operation", op)
	return nil
}

func isUnsupportedAdmissionPolicyAPI(err error) bool {
	return meta.IsNoMatchError(err)
}

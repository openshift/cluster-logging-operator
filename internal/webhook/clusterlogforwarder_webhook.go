package webhook

import (
	"context"
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	authorizationapi "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-observability-openshift-io-v1-clusterlogforwarder,mutating=false,failurePolicy=fail,sideEffects=None,groups=observability.openshift.io,resources=clusterlogforwarders,verbs=create;update,versions=v1,name=vclusterlogforwarder.observability.openshift.io,admissionReviewVersions=v1

type ClusterLogForwarderValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &ClusterLogForwarderValidator{}

func SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&obs.ClusterLogForwarder{}).
		WithValidator(&ClusterLogForwarderValidator{Client: mgr.GetClient()}).
		Complete()
}

func (v *ClusterLogForwarderValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return v.validateSAUsage(ctx, obj)
}

func (v *ClusterLogForwarderValidator) ValidateUpdate(ctx context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	return v.validateSAUsage(ctx, newObj)
}

func (v *ClusterLogForwarderValidator) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// validateSAUsage verifies the requesting user has permission to use the referenced ServiceAccount.
// This check is unconditional — any SA can be granted SCC permissions and mounted to workloads,
// so we validate usage regardless of whether outputs forward the SA token.
func (v *ClusterLogForwarderValidator) validateSAUsage(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	clf, ok := obj.(*obs.ClusterLogForwarder)
	if !ok {
		return nil, fmt.Errorf("expected ClusterLogForwarder, got %T", obj)
	}

	saName := clf.Spec.ServiceAccount.Name
	if saName == "" {
		return nil, nil
	}

	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admission request from context: %w", err)
	}

	sar := &authorizationapi.SubjectAccessReview{
		Spec: authorizationapi.SubjectAccessReviewSpec{
			User:   req.UserInfo.Username,
			Groups: req.UserInfo.Groups,
			UID:    req.UserInfo.UID,
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Group:     "",
				Verb:      "use",
				Resource:  "serviceaccounts",
				Name:      saName,
				Namespace: clf.Namespace,
			},
		},
	}

	if err := v.Client.Create(ctx, sar); err != nil {
		return nil, fmt.Errorf("failed to check service account usage permission: %w", err)
	}

	if !sar.Status.Allowed {
		return nil, fmt.Errorf("user %q is not authorized to use service account %q in namespace %q", req.UserInfo.Username, saName, clf.Namespace)
	}

	return nil, nil
}

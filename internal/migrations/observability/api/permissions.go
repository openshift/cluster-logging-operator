package api

import (
	"context"
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateLogCollectorSAPermissions(k8sClient client.Client) error {
	for _, inputType := range obs.ReservedInputTypes.List() {
		name := fmt.Sprintf("%s-collect-%s-logs", constants.CollectorServiceAccountName, inputType)
		current := &rbacv1.ClusterRoleBinding{}
		key := client.ObjectKey{Name: name}

		// Check if CRB exists
		if err := k8sClient.Get(context.TODO(), key, current); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed to get ClusterRoleBinding: %w", err)
			}
			// Create the CRB if it doesn't exist
			crb := NewLogCollectorClusterRoleBinding(name, inputType)
			if err := k8sClient.Create(context.TODO(), crb); err != nil {
				return fmt.Errorf("error  creating ClusterRoleBinding: %w", err)
			}
		}
	}

	return nil
}

func NewLogCollectorClusterRoleBinding(name, input string) *rbacv1.ClusterRoleBinding {
	return runtime.NewClusterRoleBinding(name,
		rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     fmt.Sprintf("collect-%s-logs", input),
		},
		rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      constants.CollectorServiceAccountName,
			Namespace: constants.OpenshiftNS,
		},
	)
}

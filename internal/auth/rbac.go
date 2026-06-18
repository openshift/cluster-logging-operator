package auth

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileRBAC reconciles the RBAC specifically for the service account and SCC
func ReconcileRBAC(k8sClient client.Client, rbacName, saNamespace, saName string, owner metav1.OwnerReference) error {
	desiredCRB := NewMetaDataReaderClusterRoleBinding(saNamespace, saName)
	if err := reconcile.ClusterRoleBinding(k8sClient, desiredCRB.Name, func() *rbacv1.ClusterRoleBinding { return desiredCRB }); err != nil {
		return err
	}
	desiredSCCRole := NewServiceAccountSCCRole(saNamespace, rbacName, saName, owner)
	if err := reconcile.Role(k8sClient, desiredSCCRole); err != nil {
		return err
	}

	desiredSCCRoleBinding := NewServiceAccountSCCRoleBinding(saNamespace, rbacName, saName, owner)
	if err := reconcile.RoleBinding(k8sClient, desiredSCCRoleBinding); err != nil {
		return err
	}

	// Best-effort cleanup of old resources with previous naming scheme.
	// Errors are logged but not returned to avoid blocking reconciliation
	// (e.g., namespace-scoped cache may not know about newly created namespaces).
	cleanupLegacySCCRole(k8sClient, saNamespace, saName)

	return nil
}

func cleanupLegacySCCRole(k8sClient client.Client, saNamespace, saName string) {
	oldRoleName := fmt.Sprintf("%s-scc", saName)

	roleBindings := &rbacv1.RoleBindingList{}
	if err := k8sClient.List(context.TODO(), roleBindings, client.InNamespace(saNamespace)); err != nil {
		log.V(3).Info("skipping legacy SCC role cleanup: unable to list rolebindings", "namespace", saNamespace, "error", err)
		return
	}

	for _, rb := range roleBindings.Items {
		if rb.RoleRef.Name == oldRoleName {
			return
		}
	}

	if err := reconcile.DeleteRole(k8sClient, saNamespace, oldRoleName); err != nil {
		log.V(3).Info("skipping legacy SCC role cleanup: unable to delete old role", "namespace", saNamespace, "role", oldRoleName, "error", err)
	}
}

// NewMetaDataReaderClusterRoleBinding stubs a clusterrolebinding to allow reading of pod metadata (i.e. labels)
func NewMetaDataReaderClusterRoleBinding(saNamespace, saName string) *rbacv1.ClusterRoleBinding {
	name := fmt.Sprintf("metadata-reader-%s-%s", saNamespace, saName)
	desired := runtime.NewClusterRoleBinding(name,
		rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "metadata-reader",
		},
		rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      saName,
			Namespace: saNamespace,
		},
	)

	return desired
}

func NewServiceAccountSCCRole(namespace, name, saName string, owner metav1.OwnerReference) *rbacv1.Role {
	roleName := fmt.Sprintf("%s-%s-scc", name, saName)
	sccRule := rbacv1.PolicyRule{
		APIGroups:     []string{"security.openshift.io"},
		ResourceNames: []string{sccName},
		Resources:     []string{"securitycontextconstraints"},
		Verbs:         []string{"use"},
	}

	desired := runtime.NewRole(namespace, roleName, sccRule)

	utils.AddOwnerRefToObject(desired, owner)
	return desired
}

func NewServiceAccountSCCRoleBinding(namespace, name, saName string, owner metav1.OwnerReference) *rbacv1.RoleBinding {
	roleName := fmt.Sprintf("%s-%s-scc", name, saName)
	roleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     roleName,
	}

	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      saName,
		Namespace: namespace,
	}

	roleBindingName := fmt.Sprintf("%s-scc", name)
	desired := runtime.NewRoleBinding(namespace, roleBindingName, roleRef, subject)

	utils.AddOwnerRefToObject(desired, owner)
	return desired
}

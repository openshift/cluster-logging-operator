package auth

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileRBAC reconciles the RBAC specifically for the service account and SCC
func ReconcileRBAC(k8sClient client.Client, rbacName, saNamespace, saName string, owner metav1.OwnerReference) error {
	desiredCRB := NewMetaDataReaderClusterRoleBinding(saNamespace, saName, owner)
	if err := reconcile.ClusterRoleBinding(k8sClient, desiredCRB.Name, func() *rbacv1.ClusterRoleBinding { return desiredCRB }); err != nil {
		return err
	}
	desiredSCCRole := NewServiceAccountSCCRole(saNamespace, saName, owner)
	if err := reconcile.Role(k8sClient, desiredSCCRole); err != nil {
		return err
	}

	desiredSCCRoleBinding := NewServiceAccountSCCRoleBinding(saNamespace, rbacName, desiredSCCRole.Name, saName, owner)
	return reconcile.RoleBinding(k8sClient, desiredSCCRoleBinding)
}

// NewMetaDataReaderClusterRoleBinding stubs a clusterrolebinding to allow reading of pod metadata (i.e. labels)
func NewMetaDataReaderClusterRoleBinding(saNamespace, saName string, owner metav1.OwnerReference) *rbacv1.ClusterRoleBinding {
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

	utils.AddOwnerRefToObject(desired, owner)
	return desired
}

func NewServiceAccountSCCRole(namespace, name string, owner metav1.OwnerReference) *rbacv1.Role {
	name = fmt.Sprintf("%s-scc", name)
	sccRule := rbacv1.PolicyRule{
		APIGroups:     []string{"security.openshift.io"},
		ResourceNames: []string{sccName},
		Resources:     []string{"securitycontextconstraints"},
		Verbs:         []string{"use"},
	}

	desired := runtime.NewRole(namespace, name, sccRule)

	utils.AddOwnerRefToObject(desired, owner)
	return desired
}

func NewServiceAccountSCCRoleBinding(namespace, name, roleName, saName string, owner metav1.OwnerReference) *rbacv1.RoleBinding {
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

	name = fmt.Sprintf("%s-scc", name)
	desired := runtime.NewRoleBinding(namespace, name, roleRef, subject)

	utils.AddOwnerRefToObject(desired, owner)
	return desired
}

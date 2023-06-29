package collector

import (
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileRBAC reconciles the service specifically for the collector that exposes the collector metrics
func ReconcileRBAC(er record.EventRecorder, k8sClient client.Client, saNamespace string, resNames *factory.ForwarderResourceNames, owner metav1.OwnerReference) error {
	desiredCRB := NewMetaDataReaderClusterRoleBinding(saNamespace, resNames.MetadataReaderClusterRoleBinding, resNames.ServiceAccount, owner)
	if err := reconcile.ClusterRoleBinding(k8sClient, resNames.MetadataReaderClusterRoleBinding, func() *rbacv1.ClusterRoleBinding { return desiredCRB }); err != nil {
		return err
	}
	desiredSCCRole := NewServiceAccountSCCRole(saNamespace, resNames.CommonName, owner)
	if err := reconcile.Role(er, k8sClient, desiredSCCRole); err != nil {
		return err
	}

	desiredSCCRoleBinding := NewServiceAccountSCCRoleBinding(saNamespace, resNames.CommonName, resNames.ServiceAccount, owner)
	return reconcile.RoleBinding(er, k8sClient, desiredSCCRoleBinding)
}

// NewMetaDataReaderClusterRoleBinding stubs a clusterrolebinding to allow reading of pod metadata (i.e. labels)
func NewMetaDataReaderClusterRoleBinding(saNamespace, name, saName string, owner metav1.OwnerReference) *rbacv1.ClusterRoleBinding {

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

func NewServiceAccountSCCRoleBinding(namespace, name, saName string, owner metav1.OwnerReference) *rbacv1.RoleBinding {
	roleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     name,
	}

	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      saName,
		Namespace: namespace,
	}

	desired := runtime.NewRoleBinding(namespace, name, roleRef, subject)

	utils.AddOwnerRefToObject(desired, owner)
	return desired
}

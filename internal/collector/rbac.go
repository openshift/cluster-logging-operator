package collector

import (
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileRBAC reconciles the service specifically for the collector that exposes the collector metrics
func ReconcileRBAC(er record.EventRecorder, k8sClient client.Client, saNamespace, saName string, owner metav1.OwnerReference) error {
	desired := NewMetaDataReaderClusterRoleBinding(saNamespace, "cluster-logging-metadata-reader", saName)
	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.ClusterRoleBinding(er, k8sClient, "cluster-logging-metadata-reader", func() *rbacv1.ClusterRoleBinding { return desired })
}

// NewMetaDataReaderClusterRoleBinding stubs a clusterrolebinding to allow reading of pod metadata (i.e. labels)
func NewMetaDataReaderClusterRoleBinding(saNamespace, name, saName string) *rbacv1.ClusterRoleBinding {
	return runtime.NewClusterRoleBinding(name,
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
}

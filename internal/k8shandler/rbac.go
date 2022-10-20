package k8shandler

import (
	"github.com/ViaQ/logerr/v2/kverrors"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// removeClusterRoleBinding removes a ClusterRoleBinding
func (clusterRequest *ClusterLoggingRequest) removeClusterRoleBinding(name string) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := clusterRequest.Delete(clusterRoleBinding); err != nil {
		if !apierrors.IsNotFound(err) {
			return kverrors.Wrap(err, "Failed to delete ClusterRoleBinding", "name", name)
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeClusterRole(name string) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := clusterRequest.Delete(clusterRole); err != nil {
		if !apierrors.IsNotFound(err) {
			return kverrors.Wrap(err, "Failed to delete ClusterRole", "name", name)
		}
	}

	return nil
}

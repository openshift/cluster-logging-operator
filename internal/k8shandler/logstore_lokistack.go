package k8shandler

import (
	"fmt"

	"github.com/ViaQ/logerr/v2/kverrors"
	log "github.com/ViaQ/logerr/v2/log/static"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	lokiStackClusterRoleName        = "logging-collector-logs-writer"
	lokiStackClusterRoleBindingName = "logging-collector-logs-writer"
)

func (clusterRequest *ClusterLoggingRequest) createOrUpdateLokiStackLogStore() error {
	if err := clusterRequest.createOrUpdateLokiStackClusterRole(); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for LokiStack collector.")
	}

	if err := clusterRequest.createOrUpdateLokiStackClusterRoleBinding(); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRoleBinding for LokiStack collector.")
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateLokiStackClusterRole() error {
	clusterRole := &rbacv1.ClusterRole{}
	if err := clusterRequest.Get(lokiStackClusterRoleName, clusterRole); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get ClusterRole: %w", err)
		}

		clusterRole = newLokiStackClusterRole()
		if err := clusterRequest.Create(clusterRole); err != nil {
			return fmt.Errorf("failed to create ClusterRole: %w", err)
		}

		return nil
	}

	wantRole := newLokiStackClusterRole()
	if compareLokiStackClusterRole(clusterRole, wantRole) {
		log.V(9).Info("LokiStack collector ClusterRole matches.")
		return nil
	}

	clusterRole.Rules = wantRole.Rules

	if err := clusterRequest.Update(clusterRole); err != nil {
		return fmt.Errorf("failed to update ClusterRole: %w", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateLokiStackClusterRoleBinding() error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := clusterRequest.Get(lokiStackClusterRoleBindingName, clusterRoleBinding); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get ClusterRoleBinding: %w", err)
		}

		clusterRoleBinding = newLokiStackClusterRoleBinding()
		if err := clusterRequest.Create(clusterRoleBinding); err != nil {
			return fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
		}

		return nil
	}

	wantRoleBinding := newLokiStackClusterRoleBinding()
	if compareLokiStackClusterRoleBinding(clusterRoleBinding, wantRoleBinding) {
		log.V(9).Info("LokiStack collector ClusterRoleBinding matches.")
		return nil
	}

	clusterRoleBinding.RoleRef = wantRoleBinding.RoleRef
	clusterRoleBinding.Subjects = wantRoleBinding.Subjects

	if err := clusterRequest.Update(clusterRoleBinding); err != nil {
		return fmt.Errorf("failed to update ClusterRoleBinding: %w", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeLokiStackRbac() error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackClusterRoleBindingName,
		},
	}
	if err := clusterRequest.Delete(clusterRoleBinding); err != nil {
		if !apierrors.IsNotFound(err) {
			return kverrors.Wrap(err, "Failed to delete LokiStack ClusterRoleBinding", "name", lokiStackClusterRoleBindingName)
		}
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackClusterRoleName,
		},
	}
	if err := clusterRequest.Delete(clusterRole); err != nil {
		if !apierrors.IsNotFound(err) {
			return kverrors.Wrap(err, "Failed to delete LokiStack ClusterRole", "name", lokiStackClusterRoleName)
		}
	}
	return nil
}

func newLokiStackClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"loki.grafana.com",
				},
				Resources: []string{
					"application",
					"infrastructure",
				},
				ResourceNames: []string{
					"logs",
				},
				Verbs: []string{
					"create",
				},
			},
		},
	}
}

func newLokiStackClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackClusterRoleBindingName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "logging-collector-logs-writer",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "logcollector",
				Namespace: "openshift-logging",
			},
		},
	}
}

func compareLokiStackClusterRole(got, want *rbacv1.ClusterRole) bool {
	return equality.Semantic.DeepEqual(got.Rules, want.Rules)
}

func compareLokiStackClusterRoleBinding(got, want *rbacv1.ClusterRoleBinding) bool {
	return equality.Semantic.DeepEqual(got.RoleRef, want.RoleRef) &&
		equality.Semantic.DeepEqual(got.Subjects, want.Subjects)
}

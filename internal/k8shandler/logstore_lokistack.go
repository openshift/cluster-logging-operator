package k8shandler

import (
	"github.com/ViaQ/logerr/v2/kverrors"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	lokiStackFinalizer = "logging.openshift.io/lokistack-rbac"

	lokiStackWriterClusterRoleName        = "logging-collector-logs-writer"
	lokiStackWriterClusterRoleBindingName = "logging-collector-logs-writer"

	lokiStackAppReaderClusterRoleName        = "logging-application-logs-reader"
	lokiStackAppReaderClusterRoleBindingName = "logging-all-authenticated-application-logs-reader"
)

func (clusterRequest *ClusterLoggingRequest) createOrUpdateLokiStackLogStore() error {
	if clusterRequest.Cluster.DeletionTimestamp != nil {
		// Skip creation if deleting
		return nil
	}

	if err := clusterRequest.appendFinalizer(lokiStackFinalizer); err != nil {
		return kverrors.Wrap(err, "Failed to set finalizer for LokiStack RBAC rules.")
	}

	if err := clusterRequest.createOrUpdateClusterRole(lokiStackWriterClusterRoleName, newLokiStackWriterClusterRole); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for LokiStack collector.")
	}

	if err := clusterRequest.createOrUpdateClusterRoleBinding(lokiStackWriterClusterRoleBindingName, newLokiStackWriterClusterRoleBinding); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRoleBinding for LokiStack collector.")
	}

	if err := clusterRequest.createOrUpdateClusterRole(lokiStackAppReaderClusterRoleName, newLokiStackAppReaderClusterRole); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for reading application logs.")
	}

	if err := clusterRequest.createOrUpdateClusterRoleBinding(lokiStackAppReaderClusterRoleBindingName, newLokiStackAppReaderClusterRoleBinding); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRoleBinding for reading application logs.")
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeLokiStackRbac() error {
	if err := clusterRequest.removeClusterRoleBinding(lokiStackAppReaderClusterRoleBindingName); err != nil {
		return err
	}

	if err := clusterRequest.removeClusterRole(lokiStackAppReaderClusterRoleName); err != nil {
		return err
	}

	if err := clusterRequest.removeClusterRoleBinding(lokiStackWriterClusterRoleBindingName); err != nil {
		return err
	}

	if err := clusterRequest.removeClusterRole(lokiStackWriterClusterRoleName); err != nil {
		return err
	}

	if err := clusterRequest.removeFinalizer(lokiStackFinalizer); err != nil {
		return err
	}

	return nil
}

func newLokiStackWriterClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackWriterClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"loki.grafana.com",
				},
				Resources: []string{
					"application",
					"audit",
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

func newLokiStackWriterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackWriterClusterRoleBindingName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     lokiStackWriterClusterRoleName,
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

func newLokiStackAppReaderClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackAppReaderClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"loki.grafana.com",
				},
				Resources: []string{
					"application",
				},
				ResourceNames: []string{
					"logs",
				},
				Verbs: []string{
					"get",
				},
			},
		},
	}
}

func newLokiStackAppReaderClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: lokiStackAppReaderClusterRoleBindingName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     lokiStackAppReaderClusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     "system:authenticated",
			},
		},
	}
}

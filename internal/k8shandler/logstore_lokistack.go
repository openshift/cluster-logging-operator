package k8shandler

import (
	"github.com/ViaQ/logerr/v2/kverrors"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
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

	if err := reconcile.ClusterRole(clusterRequest.Client, lokiStackWriterClusterRoleName, newLokiStackWriterClusterRole); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for LokiStack collector.")
	}

	if err := reconcile.ClusterRoleBinding(clusterRequest.Client, lokiStackWriterClusterRoleBindingName, newLokiStackWriterClusterRoleBinding); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRoleBinding for LokiStack collector.")
	}

	if err := reconcile.ClusterRole(clusterRequest.Client, lokiStackAppReaderClusterRoleName, newLokiStackAppReaderClusterRole); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for reading application logs.")
	}

	if err := reconcile.ClusterRoleBinding(clusterRequest.Client, lokiStackAppReaderClusterRoleBindingName, newLokiStackAppReaderClusterRoleBinding); err != nil {
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

	return runtime.NewClusterRole(lokiStackWriterClusterRoleName,
		rbacv1.PolicyRule{
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
	)
}

func newLokiStackWriterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return runtime.NewClusterRoleBinding(
		lokiStackWriterClusterRoleBindingName,
		rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     lokiStackWriterClusterRoleName,
		}, rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      "logcollector",
			Namespace: "openshift-logging",
		},
	)
}

func newLokiStackAppReaderClusterRole() *rbacv1.ClusterRole {
	return runtime.NewClusterRole(
		lokiStackAppReaderClusterRoleName,
		rbacv1.PolicyRule{
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
	)
}

func newLokiStackAppReaderClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return runtime.NewClusterRoleBinding(
		lokiStackAppReaderClusterRoleBindingName,
		rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     lokiStackAppReaderClusterRoleName,
		},
		rbacv1.Subject{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Group",
			Name:     "system:authenticated",
		},
	)
}

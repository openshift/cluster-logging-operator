package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"sort"
	"strings"

	"github.com/ViaQ/logerr/v2/kverrors"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/utils/strings/slices"
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

func (clusterRequest *ClusterLoggingRequest) processPipelinesForLokiStack(inPipelines []loggingv1.PipelineSpec) ([]loggingv1.OutputSpec, []loggingv1.PipelineSpec) {
	needOutput := make(map[string]bool)
	pipelines := []loggingv1.PipelineSpec{}

	for _, p := range inPipelines {
		if !slices.Contains(p.OutputRefs, loggingv1.OutputNameDefault) {
			// Skip pipelines that do not reference "default" output
			pipelines = append(pipelines, p)
			continue
		}

		for _, i := range p.InputRefs {
			needOutput[i] = true
		}

		for i, input := range p.InputRefs {
			pOut := p.DeepCopy()
			pOut.InputRefs = []string{input}

			for i, output := range pOut.OutputRefs {
				if output != loggingv1.OutputNameDefault {
					// Leave non-default output names as-is
					continue
				}

				pOut.OutputRefs[i] = lokiStackOutput(input)
			}

			if pOut.Name != "" && i > 0 {
				// Generate new name for named pipelines as duplicate names are not allowed
				pOut.Name = fmt.Sprintf("%s-%d", pOut.Name, i)
			}

			pipelines = append(pipelines, *pOut)
		}
	}

	outputs := []loggingv1.OutputSpec{}
	for input := range needOutput {
		outputs = append(outputs, loggingv1.OutputSpec{
			Name: lokiStackOutput(input),
			Type: loggingv1.OutputTypeLoki,
			URL:  clusterRequest.LokiStackURL(input),
		})
	}

	// Sort outputs, because we have tests depending on the exact generated configuration
	sort.Slice(outputs, func(i, j int) bool {
		return strings.Compare(outputs[i].Name, outputs[j].Name) < 0
	})

	return outputs, pipelines
}

func lokiStackOutput(inputName string) string {
	switch inputName {
	case loggingv1.InputNameApplication:
		return loggingv1.OutputNameDefault + "-loki-apps"
	case loggingv1.InputNameInfrastructure:
		return loggingv1.OutputNameDefault + "-loki-infra"
	case loggingv1.InputNameAudit:
		return loggingv1.OutputNameDefault + "-loki-audit"
	}

	return ""
}

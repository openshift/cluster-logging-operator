package k8shandler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ViaQ/logerr/v2/kverrors"
	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
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

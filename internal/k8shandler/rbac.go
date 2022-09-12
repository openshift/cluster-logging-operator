package k8shandler

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/ViaQ/logerr/v2/kverrors"
	log "github.com/ViaQ/logerr/v2/log/static"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewPolicyRule stubs policy rule
func NewPolicyRule(apiGroups, resources, resourceNames, verbs []string) rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups:     apiGroups,
		Resources:     resources,
		ResourceNames: resourceNames,
		Verbs:         verbs,
	}
}

//NewPolicyRules stubs policy rules
func NewPolicyRules(rules ...rbacv1.PolicyRule) []rbacv1.PolicyRule {
	return rules
}

// NewRole stubs a role
func NewRole(roleName, namespace string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
		Rules: rules,
	}
}

// NewSubject stubs a new subject
func NewSubject(kind, name string) rbacv1.Subject {
	return rbacv1.Subject{
		Kind:     kind,
		Name:     name,
		APIGroup: rbacv1.GroupName,
	}
}

// NewSubjects stubs subjects
func NewSubjects(subjects ...rbacv1.Subject) []rbacv1.Subject {
	return subjects
}

// NewRoleBinding stubs a role binding
func NewRoleBinding(bindingName, namespace, roleName string, subjects []rbacv1.Subject) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: rbacv1.GroupName,
		},
		Subjects: subjects,
	}
}

// NewClusterRoleBinding stubs a cluster role binding
func NewClusterRoleBinding(bindingName, roleName string, subjects []rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     roleName,
			APIGroup: rbacv1.GroupName,
		},
		Subjects: subjects,
	}
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateClusterRole(name string, generator func() *rbacv1.ClusterRole) error {
	clusterRole := &rbacv1.ClusterRole{}
	if err := clusterRequest.Get(name, clusterRole); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get ClusterRole: %w", err)
		}

		clusterRole = generator()
		if err := clusterRequest.Create(clusterRole); err != nil {
			return fmt.Errorf("failed to create ClusterRole: %w", err)
		}

		return nil
	}

	wantRole := generator()
	if compareClusterRole(clusterRole, wantRole) {
		log.V(9).Info("LokiStack collector ClusterRole matches.")
		return nil
	}

	clusterRole.Rules = wantRole.Rules

	if err := clusterRequest.Update(clusterRole); err != nil {
		return fmt.Errorf("failed to update ClusterRole: %w", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateClusterRoleBinding(name string, generator func() *rbacv1.ClusterRoleBinding) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := clusterRequest.Get(name, clusterRoleBinding); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get ClusterRoleBinding: %w", err)
		}

		clusterRoleBinding = generator()
		if err := clusterRequest.Create(clusterRoleBinding); err != nil {
			return fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
		}

		return nil
	}

	wantRoleBinding := generator()
	if compareClusterRoleBinding(clusterRoleBinding, wantRoleBinding) {
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

func compareClusterRole(got, want *rbacv1.ClusterRole) bool {
	return equality.Semantic.DeepEqual(got.Rules, want.Rules)
}

func compareClusterRoleBinding(got, want *rbacv1.ClusterRoleBinding) bool {
	return equality.Semantic.DeepEqual(got.RoleRef, want.RoleRef) &&
		equality.Semantic.DeepEqual(got.Subjects, want.Subjects)
}

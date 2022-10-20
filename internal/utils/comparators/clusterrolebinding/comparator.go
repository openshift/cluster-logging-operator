package clusterrolebinding

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func AreSame(got, want *rbacv1.ClusterRoleBinding) bool {
	return equality.Semantic.DeepEqual(got.RoleRef, want.RoleRef) &&
		equality.Semantic.DeepEqual(got.Subjects, want.Subjects)
}

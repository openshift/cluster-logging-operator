package rolebinding

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func AreSame(got, want *rbacv1.RoleBinding) bool {
	return equality.Semantic.DeepEqual(got.RoleRef, want.RoleRef) &&
		equality.Semantic.DeepEqual(got.Subjects, want.Subjects)
}

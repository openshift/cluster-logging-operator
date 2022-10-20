package clusterrole

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func AreSame(got, want *rbacv1.ClusterRole) bool {
	return equality.Semantic.DeepEqual(got.Rules, want.Rules)
}

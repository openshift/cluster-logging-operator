package clusterrolebinding_test

import (
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/clusterrolebinding"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestCompareClusterRoleBinding(t *testing.T) {
	roleRef := rbacv1.RoleRef{
		APIGroup: "apigroup",
		Kind:     "kind",
		Name:     "name",
	}

	subject := rbacv1.Subject{
		Kind:      "kind",
		APIGroup:  "apigroup",
		Name:      "name",
		Namespace: "namespace",
	}

	tests := []struct {
		name        string
		gotBinding  *rbacv1.ClusterRoleBinding
		wantBinding *rbacv1.ClusterRoleBinding
		wantCompare bool
	}{
		{
			name:        "empty",
			gotBinding:  &rbacv1.ClusterRoleBinding{},
			wantBinding: &rbacv1.ClusterRoleBinding{},
			wantCompare: true,
		},
		{
			name: "same RoleRef",
			gotBinding: &rbacv1.ClusterRoleBinding{
				RoleRef: roleRef,
			},
			wantBinding: &rbacv1.ClusterRoleBinding{
				RoleRef: roleRef,
			},
			wantCompare: true,
		},
		{
			name:       "different RoleRef",
			gotBinding: &rbacv1.ClusterRoleBinding{},
			wantBinding: &rbacv1.ClusterRoleBinding{
				RoleRef: roleRef,
			},
			wantCompare: false,
		},
		{
			name: "same Subjects",
			gotBinding: &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantBinding: &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantCompare: true,
		},
		{
			name:       "different Subjects",
			gotBinding: &rbacv1.ClusterRoleBinding{},
			wantBinding: &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantCompare: false,
		},
		{
			name: "different Subject",
			gotBinding: &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					{},
				},
			},
			wantBinding: &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantCompare: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clusterrolebinding.AreSame(tt.gotBinding, tt.wantBinding); got != tt.wantCompare {
				t.Errorf("CompareClusterRoleBinding() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

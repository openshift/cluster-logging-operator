package rolebinding_test

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/rolebinding"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestCompareRoleBinding(t *testing.T) {
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
		gotBinding  *rbacv1.RoleBinding
		wantBinding *rbacv1.RoleBinding
		wantCompare bool
	}{
		{
			name:        "empty",
			gotBinding:  &rbacv1.RoleBinding{},
			wantBinding: &rbacv1.RoleBinding{},
			wantCompare: true,
		},
		{
			name: "same RoleRef",
			gotBinding: &rbacv1.RoleBinding{
				RoleRef: roleRef,
			},
			wantBinding: &rbacv1.RoleBinding{
				RoleRef: roleRef,
			},
			wantCompare: true,
		},
		{
			name:       "different RoleRef",
			gotBinding: &rbacv1.RoleBinding{},
			wantBinding: &rbacv1.RoleBinding{
				RoleRef: roleRef,
			},
			wantCompare: false,
		},
		{
			name: "same Subjects",
			gotBinding: &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantBinding: &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantCompare: true,
		},
		{
			name:       "different Subjects",
			gotBinding: &rbacv1.RoleBinding{},
			wantBinding: &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantCompare: false,
		},
		{
			name: "different Subject",
			gotBinding: &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					{},
				},
			},
			wantBinding: &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					subject,
				},
			},
			wantCompare: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rolebinding.AreSame(tt.gotBinding, tt.wantBinding); got != tt.wantCompare {
				t.Errorf("CompareRoleBinding() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

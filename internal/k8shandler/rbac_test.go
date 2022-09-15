package k8shandler

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestCompareLokiStackClusterRole(t *testing.T) {
	rules := []rbacv1.PolicyRule{
		{
			Verbs:           []string{"verb"},
			APIGroups:       []string{"apigroup"},
			Resources:       []string{"resource"},
			ResourceNames:   []string{"resourcename"},
			NonResourceURLs: []string{"url"},
		},
	}

	tests := []struct {
		name        string
		gotRole     *rbacv1.ClusterRole
		wantRole    *rbacv1.ClusterRole
		wantCompare bool
	}{
		{
			name:        "empty",
			gotRole:     &rbacv1.ClusterRole{},
			wantRole:    &rbacv1.ClusterRole{},
			wantCompare: true,
		},
		{
			name: "same rules",
			gotRole: &rbacv1.ClusterRole{
				Rules: rules,
			},
			wantRole: &rbacv1.ClusterRole{
				Rules: rules,
			},
			wantCompare: true,
		},
		{
			name:    "different rules",
			gotRole: &rbacv1.ClusterRole{},
			wantRole: &rbacv1.ClusterRole{
				Rules: rules,
			},
			wantCompare: false,
		},
		{
			name: "different rule",
			gotRole: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{
					{},
				},
			},
			wantRole: &rbacv1.ClusterRole{
				Rules: rules,
			},
			wantCompare: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareClusterRole(tt.gotRole, tt.wantRole); got != tt.wantCompare {
				t.Errorf("compareClusterRole() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

func TestCompareLokiStackClusterRoleBinding(t *testing.T) {
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
			if got := compareClusterRoleBinding(tt.gotBinding, tt.wantBinding); got != tt.wantCompare {
				t.Errorf("compareClusterRoleBinding() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

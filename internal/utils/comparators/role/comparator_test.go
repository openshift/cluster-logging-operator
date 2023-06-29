package role_test

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/role"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestCompareRole(t *testing.T) {
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
		gotRole     *rbacv1.Role
		wantRole    *rbacv1.Role
		wantCompare bool
	}{
		{
			name:        "empty",
			gotRole:     &rbacv1.Role{},
			wantRole:    &rbacv1.Role{},
			wantCompare: true,
		},
		{
			name: "same rules",
			gotRole: &rbacv1.Role{
				Rules: rules,
			},
			wantRole: &rbacv1.Role{
				Rules: rules,
			},
			wantCompare: true,
		},
		{
			name:    "different rules",
			gotRole: &rbacv1.Role{},
			wantRole: &rbacv1.Role{
				Rules: rules,
			},
			wantCompare: false,
		},
		{
			name: "different rule",
			gotRole: &rbacv1.Role{
				Rules: []rbacv1.PolicyRule{
					{},
				},
			},
			wantRole: &rbacv1.Role{
				Rules: rules,
			},
			wantCompare: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := role.AreSame(tt.gotRole, tt.wantRole); got != tt.wantCompare {
				t.Errorf("CompareRole() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

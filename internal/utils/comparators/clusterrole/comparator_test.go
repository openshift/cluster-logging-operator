package clusterrole_test

import (
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/clusterrole"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestClusterRole(t *testing.T) {
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
			if got := clusterrole.AreSame(tt.gotRole, tt.wantRole); got != tt.wantCompare {
				t.Errorf("CompareClusterRole() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

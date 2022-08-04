package k8shandler

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			if got := compareLokiStackClusterRole(tt.gotRole, tt.wantRole); got != tt.wantCompare {
				t.Errorf("compareLokiStackClusterRole() = %v, want %v", got, tt.wantCompare)
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
			if got := compareLokiStackClusterRoleBinding(tt.gotBinding, tt.wantBinding); got != tt.wantCompare {
				t.Errorf("compareLokiStackClusterRoleBinding() = %v, want %v", got, tt.wantCompare)
			}
		})
	}
}

func TestProcessPipelinesForLokiStack(t *testing.T) {
	tests := []struct {
		desc          string
		in            []loggingv1.PipelineSpec
		wantOutputs   []loggingv1.OutputSpec
		wantPipelines []loggingv1.PipelineSpec
	}{
		{
			desc: "no default output",
			in: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{"custom-output"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{"custom-output"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			},
		},
		{
			desc: "single tenant - single output",
			in: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{loggingv1.OutputNameDefault},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{
				{
					Name: loggingv1.OutputNameDefault + "-loki-apps",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/application",
				},
			},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{loggingv1.OutputNameDefault + "-loki-apps"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			},
		},
		{
			desc: "multiple tenants - single output",
			in: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{loggingv1.OutputNameDefault},
					InputRefs: []string{
						loggingv1.InputNameApplication,
						loggingv1.InputNameInfrastructure,
					},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{
				{
					Name: loggingv1.OutputNameDefault + "-loki-apps",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/application",
				},
				{
					Name: loggingv1.OutputNameDefault + "-loki-infra",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/infrastructure",
				},
			},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{loggingv1.OutputNameDefault + "-loki-apps"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
				{
					OutputRefs: []string{loggingv1.OutputNameDefault + "-loki-infra"},
					InputRefs:  []string{loggingv1.InputNameInfrastructure},
				},
			},
		},
		{
			desc: "multiple tenants - single output - named pipeline",
			in: []loggingv1.PipelineSpec{
				{
					Name:       "named-pipeline",
					OutputRefs: []string{loggingv1.OutputNameDefault},
					InputRefs: []string{
						loggingv1.InputNameApplication,
						loggingv1.InputNameInfrastructure,
					},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{
				{
					Name: loggingv1.OutputNameDefault + "-loki-apps",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/application",
				},
				{
					Name: loggingv1.OutputNameDefault + "-loki-infra",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/infrastructure",
				},
			},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					Name:       "named-pipeline",
					OutputRefs: []string{loggingv1.OutputNameDefault + "-loki-apps"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
				{
					Name:       "named-pipeline-1",
					OutputRefs: []string{loggingv1.OutputNameDefault + "-loki-infra"},
					InputRefs:  []string{loggingv1.InputNameInfrastructure},
				},
			},
		},
		{
			desc: "single tenant - multiple outputs",
			in: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{
						"custom-output",
						loggingv1.OutputNameDefault,
					},
					InputRefs: []string{
						loggingv1.InputNameInfrastructure,
					},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{
				{
					Name: loggingv1.OutputNameDefault + "-loki-infra",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/infrastructure",
				},
			},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{
						"custom-output",
						loggingv1.OutputNameDefault + "-loki-infra",
					},
					InputRefs: []string{loggingv1.InputNameInfrastructure},
				},
			},
		},
		{
			desc: "multiple tenants - multiple outputs",
			in: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{
						"custom-output",
						loggingv1.OutputNameDefault,
					},
					InputRefs: []string{
						loggingv1.InputNameInfrastructure,
						loggingv1.InputNameAudit,
					},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{
				{
					Name: loggingv1.OutputNameDefault + "-loki-audit",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/audit",
				},
				{
					Name: loggingv1.OutputNameDefault + "-loki-infra",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/infrastructure",
				},
			},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{
						"custom-output",
						loggingv1.OutputNameDefault + "-loki-infra",
					},
					InputRefs: []string{loggingv1.InputNameInfrastructure},
				},
				{
					OutputRefs: []string{
						"custom-output",
						loggingv1.OutputNameDefault + "-loki-audit",
					},
					InputRefs: []string{loggingv1.InputNameAudit},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			cr := &ClusterLoggingRequest{
				Cluster: &loggingv1.ClusterLogging{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: aNamespace,
					},
					Spec: loggingv1.ClusterLoggingSpec{
						LogStore: &loggingv1.LogStoreSpec{
							Type: loggingv1.LogStoreTypeLokiStack,
							LokiStack: loggingv1.LokiStackStoreSpec{
								Name: "lokistack-testing",
							},
						},
					},
				},
			}
			outputs, pipelines := cr.processPipelinesForLokiStack(tc.in)

			if diff := cmp.Diff(outputs, tc.wantOutputs); diff != "" {
				t.Errorf("outputs differ: -got+want\n%s", diff)
			}

			if diff := cmp.Diff(pipelines, tc.wantPipelines); diff != "" {
				t.Errorf("pipelines differ: -got+want\n%s", diff)
			}
		})
	}
}

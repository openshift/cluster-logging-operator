package migrations

import (
	"github.com/google/go-cmp/cmp"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"testing"
)

func TestProcessPipelinesForLokiStack(t *testing.T) {
	tests := []struct {
		desc          string
		spec          loggingv1.ClusterLogForwarderSpec
		wantOutputs   []loggingv1.OutputSpec
		wantPipelines []loggingv1.PipelineSpec
	}{
		{
			desc: "no default output",
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
					{
						OutputRefs: []string{"custom-output"},
						InputRefs:  []string{loggingv1.InputNameApplication},
					},
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
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
					{
						OutputRefs: []string{loggingv1.OutputNameDefault},
						InputRefs:  []string{loggingv1.InputNameApplication},
					},
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
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
					{
						OutputRefs: []string{loggingv1.OutputNameDefault},
						InputRefs: []string{
							loggingv1.InputNameApplication,
							loggingv1.InputNameInfrastructure,
						},
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
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name:       "named-pipeline",
						OutputRefs: []string{loggingv1.OutputNameDefault},
						InputRefs: []string{
							loggingv1.InputNameApplication,
							loggingv1.InputNameInfrastructure,
						},
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
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
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
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
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
		{
			desc: "single tenant - custom input - default output",
			spec: loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{
					{
						OutputRefs: []string{loggingv1.OutputNameDefault},
						InputRefs:  []string{"my-custom-input"},
					},
				},
				Inputs: []loggingv1.InputSpec{
					{
						Name: "my-custom-input",
						Application: &loggingv1.Application{
							Namespaces: []string{"my-test-namespace"},
						},
					},
				},
			},
			wantOutputs: []loggingv1.OutputSpec{
				{
					Name: loggingv1.OutputNameDefault + "-my-custom-input",
					Type: loggingv1.OutputTypeLoki,
					URL:  "https://lokistack-testing-gateway-http.aNamespace.svc:8080/api/logs/v1/application",
				},
			},
			wantPipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{loggingv1.OutputNameDefault + "-my-custom-input"},
					InputRefs:  []string{"my-custom-input"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			logStore := &loggingv1.LogStoreSpec{
				Type: loggingv1.LogStoreTypeLokiStack,
				LokiStack: loggingv1.LokiStackStoreSpec{
					Name: "lokistack-testing",
				},
			}
			outputs, pipelines := processPipelinesForLokiStack(logStore, "aNamespace", tc.spec)

			if diff := cmp.Diff(outputs, tc.wantOutputs); diff != "" {
				t.Errorf("outputs differ: -got+want\n%s", diff)
			}

			if diff := cmp.Diff(pipelines, tc.wantPipelines); diff != "" {
				t.Errorf("pipelines differ: -got+want\n%s", diff)
			}
		})
	}
}

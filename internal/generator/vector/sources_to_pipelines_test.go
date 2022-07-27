package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/openshift/cluster-logging-operator/test/helpers"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			Inputs(&clfspec, op),
			Pipelines(&clfspec, op),
		)
	}
	DescribeTable("Source(s) to Pipeline(s)", helpers.TestGenerateConfWith(f),
		Entry("Send all log types to output by name", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.pipeline_user_defined]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''
`,
		}),
		Entry("Send same logtype to multiple output", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline1",
					},
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline2",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.pipeline1_user_defined]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''

[transforms.pipeline2_user_defined]
type = "remap"
inputs = ["application"]
source = '''
  .
'''
`,
		}),
		Entry("Route Logs by Namespace(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"test-ns1", "test-ns2"},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"myapplogs"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.myapplogs = '(.kubernetes.namespace_name == "test-ns1") || (.kubernetes.namespace_name == "test-ns2")'

[transforms.pipeline_user_defined]
type = "remap"
inputs = ["route_application_logs.myapplogs"]
source = '''
  .
'''
`,
		}),
		Entry("Route Logs by Namespaces(s), and Labels(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"myapp1", "myapp2"},
							Selector: &logging.LabelSelector{
								MatchLabels: map[string]string{
									"key1": "value1",
									"key2": "value2",
								},
							},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"myapplogs"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.myapplogs = '((.kubernetes.namespace_name == "myapp1") || (.kubernetes.namespace_name == "myapp2")) && ((.kubernetes.labels."key1" == "value1") && (.kubernetes.labels."key2" == "value2"))'

[transforms.pipeline_user_defined]
type = "remap"
inputs = ["route_application_logs.myapplogs"]
source = '''
  .
'''
`,
		}),
		Entry("Add Openshift Label(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

[transforms.pipeline_user_defined]
type = "remap"
inputs = ["application","infrastructure"]
source = '''
  .openshift.labels = {"label1":"value1"}
'''
`,
		}),
		Entry("Parse log message as Json", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
						Parse:      "json",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

[transforms.pipeline_user_defined]
type = "remap"
inputs = ["application","infrastructure"]
source = '''
  if .log_type == "application" {
    parsed, err = parse_json(.message)
      if err == null {
        .structured = parsed
        del(.message)
    }
  }
'''
`,
		}),
		Entry("Detect Multi Line Exceptions", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:             []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						OutputRefs:            []string{logging.OutputNameDefault},
						Name:                  "pipeline",
						DetectMultilineErrors: true,
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

[transforms.detect_exceptions_pipeline_user_defined]
type = "detect_exceptions"
inputs = ["application","infrastructure"]
languages = ["All"]
group_by = ["kubernetes.namespace_name","kubernetes.pod_name","kubernetes.container_name", "kubernetes.pod_id"]
expire_after_ms = 2000
multiline_flush_interval_ms = 1000

[transforms.pipeline_user_defined]
type = "remap"
inputs = ["detect_exceptions_pipeline_user_defined"]
source = '''
  .
'''
`,
		}),
		Entry("pipeline with spaces are sanitized", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline with space",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.pipeline_with_space_user_defined]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''
`,
		}),
		Entry("Application Inputs with Container Limits", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "input-app1",
						Application: &logging.Application{
							Namespaces: []string{"logstress"},
							Selector: &logging.LabelSelector{
								MatchLabels: map[string]string{
									"podname": "very-important",
								},
							},
							ContainerLimit: &logging.LimitSpec{
								MaxRecordsPerSecond: 100,
							},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"input-app1"},
						OutputRefs: []string{"default"},
						Name:       "flow-control",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
	.log_type = "application"
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.input-app1 = '(.kubernetes.namespace_name == "logstress") && (.kubernetes.labels."podname" == "very-important")'


[transforms.source_throttle_input-app1]
type = "throttle"
inputs = ["route_application_logs.input-app1"]
window_secs = 1
threshold = 100
key_field = "{{ file }}"

[transforms.flow_control_user_defined]
type = "remap"
inputs = ["source_throttle_input-app1"]
source = '''
	.
'''
`,
		}),
		Entry("Application Inputs with Group Limits", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "input-app2",
						Application: &logging.Application{
							Namespaces: []string{"logstress"},
							Selector: &logging.LabelSelector{
								MatchLabels: map[string]string{
									"podname": "less-important",
								},
							},
							GroupLimit: &logging.LimitSpec{
								MaxRecordsPerSecond: 50,
							},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"input-app2"},
						OutputRefs: []string{"default"},
						Name:       "flow-control",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
	.log_type = "application"
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.input-app2 = '(.kubernetes.namespace_name == "logstress") && (.kubernetes.labels."podname" == "less-important")'


[transforms.source_throttle_input-app2]
type = "throttle"
inputs = ["route_application_logs.input-app2"]
window_secs = 1
threshold = 50

[transforms.flow_control_user_defined]
type = "remap"
inputs = ["source_throttle_input-app2"]
source = '''
	.
'''
`,
		}),
	)
})

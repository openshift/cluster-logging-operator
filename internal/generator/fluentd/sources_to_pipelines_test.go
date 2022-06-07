package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/**
Note: This test case includes only the dynamic parts of Fluentd conf. This leaves out the static parts which do not change with CLF spec.
**/
var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			SourcesToInputs(&clfspec, op),
			InputsToPipeline(&clfspec, op),
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
# Include Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type relabel
  @label @_INFRASTRUCTURE
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Include Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type relabel
  @label @_AUDIT
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Sending application source type to pipeline
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @PIPELINE
  </match>
</label>

# Sending infrastructure source type to pipeline
<label @_INFRASTRUCTURE>
  <filter **>
    @type record_modifier
    <record>
      log_type infrastructure
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @PIPELINE
  </match>
</label>

# Sending audit source type to pipeline
<label @_AUDIT>
  <filter **>
    @type record_modifier
    <record>
      log_type audit
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @PIPELINE
  </match>
</label>`,
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
# Include Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type relabel
  @label @_INFRASTRUCTURE
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Include Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type relabel
  @label @_AUDIT
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Copying application source type to pipeline
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type copy
    <store>
      @type relabel
      @label @PIPELINE1
    </store>
    
    <store>
      @type relabel
      @label @PIPELINE2
    </store>
  </match>
</label>

# Sending infrastructure source type to pipeline
<label @_INFRASTRUCTURE>
  <filter **>
    @type record_modifier
    <record>
      log_type infrastructure
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @PIPELINE1
  </match>
</label>

# Sending audit source type to pipeline
<label @_AUDIT>
  <filter **>
    @type record_modifier
    <record>
      log_type audit
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @PIPELINE1
  </match>
</label>
`,
		}),
		Entry("Route Logs by Namespace(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"myapp1", "myapp2"},
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
# Discard Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type null
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Discard Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type null
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @PIPELINE
      <match>
        namespaces myapp1, myapp2
      </match>
    </route>
  </match>
</label>`,
		}),
		Entry("Route Logs by Labels(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Selector: &v1.LabelSelector{
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
# Discard Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type null
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Discard Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type null
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @PIPELINE
      <match>
        labels key1:value1, key2:value2
      </match>
    </route>
  </match>
</label>`,
		}),
		Entry("Route Logs by Namespaces(s), and Labels(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"myapp1", "myapp2"},
							Selector: &v1.LabelSelector{
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
# Discard Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type null
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Discard Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type null
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @PIPELINE
      <match>
        namespaces myapp1, myapp2
        labels key1:value1, key2:value2
      </match>
    </route>
  </match>
</label>`,
		}),
		Entry("Send Logs by custom selection, and direct", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"myapp1", "myapp2"},
							Selector: &v1.LabelSelector{
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
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline1",
					},
					{
						InputRefs:  []string{"myapplogs"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline2",
					},
				},
			},
			ExpectedConf: `
# Discard Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type null
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Discard Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type null
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @PIPELINE2
      <match>
        namespaces myapp1, myapp2
        labels key1:value1, key2:value2
      </match>
    </route>
    
    <route>
      @label @_APPLICATION_ALL
      <match>
      
      </match>
    </route>
  </match>
</label>

# Sending unrouted application to pipelines
<label @_APPLICATION_ALL>
  <match **>
    @type relabel
    @label @PIPELINE1
  </match>
</label>`,
		}),
		Entry("Complex Case", helpers.ConfGenerateTest{
			Desc: "Complex Case",
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name:        "myapplogs1",
						Application: &logging.Application{},
					},
					{
						Name: "myapplogs2",
						Application: &logging.Application{
							Namespaces: []string{"myapp"},
							Selector: &v1.LabelSelector{
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
						InputRefs:  []string{"myapplogs1"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline1",
					},
					{
						InputRefs:  []string{"myapplogs2"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline2",
					},
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline3",
					},
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline4",
					},
				},
			},
			ExpectedConf: `
# Discard Infrastructure logs
<match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
  @type null
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Discard Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type null
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @PIPELINE2
      <match>
        namespaces myapp
        labels key1:value1, key2:value2
      </match>
    </route>
    
    <route>
      @label @_APPLICATION_ALL
      <match>
      
      </match>
    </route>
  </match>
</label>

# Copying unrouted application to pipelines
<label @_APPLICATION_ALL>
  <match **>
    @type copy
    <store>
      @type relabel
      @label @PIPELINE1
    </store>
    
    <store>
      @type relabel
      @label @PIPELINE3
    </store>
    
    <store>
      @type relabel
      @label @PIPELINE4
    </store>
  </match>
</label>`,
		}),
	)
})

package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return PipelineToOutputs(&clfspec, generator.NoOptions)
	}
	DescribeTable("Pipelines(s) to Output(s)", helpers.TestGenerateConfWith(f),
		Entry("Application to single output", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "defaultoutput",
					},
				},
			},
			ExpectedConf: `
# Copying pipeline defaultoutput to outputs
<label @DEFAULTOUTPUT>
  <match **>
    @type relabel
    @label @DEFAULT
  </match>
</label>`,
		}),
		Entry("Application to multiple outputs", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault, "es-app-out"},
						Name:       "app-to-es",
					},
					{
						InputRefs:  []string{logging.InputNameAudit},
						OutputRefs: []string{logging.OutputNameDefault, "es-audit-out"},
						Name:       "audit-to-es",
					},
				},
			},
			ExpectedConf: `
# Copying pipeline app-to-es to outputs
<label @APP_TO_ES>
  <match **>
    @type copy
    deep_copy true
    <store>
      @type relabel
      @label @DEFAULT
    </store>
    
    <store>
      @type relabel
      @label @ES_APP_OUT
    </store>
  </match>
</label>

# Copying pipeline audit-to-es to outputs
<label @AUDIT_TO_ES>
  <match **>
    @type copy
    deep_copy true
    <store>
      @type relabel
      @label @DEFAULT
    </store>
    
    <store>
      @type relabel
      @label @ES_AUDIT_OUT
    </store>
  </match>
</label>`,
		}),
		Entry("Application to default output with Labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault, "es-app-out"},
						Name:       "app-to-es",
						Labels: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
			ExpectedConf: `
# Copying pipeline app-to-es to outputs
<label @APP_TO_ES>
  # Add User Defined labels to the output record
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"a":"b","c":"d"} }
    </record>
  </filter>
  
  <match **>
    @type copy
    deep_copy true
    <store>
      @type relabel
      @label @DEFAULT
    </store>
    
    <store>
      @type relabel
      @label @ES_APP_OUT
    </store>
  </match>
</label>`,
		}),
		Entry("Application to default output with Json Parsing, and Labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault, "es-app-out"},
						Name:       "app-to-es",
						Parse:      "json",
						Labels: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
			ExpectedConf: `
# Copying pipeline app-to-es to outputs
<label @APP_TO_ES>
  # Add User Defined labels to the output record
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"a":"b","c":"d"} }
    </record>
  </filter>
  
  # Parse the logs into json
  <filter **>
    @type parser
    key_name message
    reserve_data yes
    hash_value_field structured
    <parse>
      @type json
      json_parser oj
    </parse>
  </filter>
  
  <match **>
    @type copy
    deep_copy true
    <store>
      @type relabel
      @label @DEFAULT
    </store>
    
    <store>
      @type relabel
      @label @ES_APP_OUT
    </store>
  </match>
</label>`,
		}),
	)
})

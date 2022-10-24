package fluentd

import (
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

var _ = Describe("Generating pipeline to output labels", func() {
	var (
		g generator.Generator
	)
	BeforeEach(func() {
		g = generator.MakeGenerator()
	})

	It("should generate no labels for a single pipeline", func() {
		pipelines := []logging.PipelineSpec{
			{
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"infra-es"},
				Labels:     nil,
				Name:       "Pipeline-1",
			},
		}
		c := PipelineToOutputs(&logging.ClusterLogForwarderSpec{
			Pipelines: pipelines,
		}, nil)
		got, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(got).To(EqualTrimLines(`
# Copying pipeline Pipeline-1 to outputs
<label @PIPELINE_1>
  <match **>
    @type relabel
    @label @INFRA_ES
  </match>
</label>
`))
	})

	It("should generate one label for a single pipeline", func() {
		pipelines := []logging.PipelineSpec{
			{
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"infra-es"},
				Labels:     map[string]string{"1": "2"},
				Name:       "Pipeline-1",
			},
		}
		c := PipelineToOutputs(&logging.ClusterLogForwarderSpec{
			Pipelines: pipelines,
		}, nil)
		got, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(got).To(EqualTrimLines(`
# Copying pipeline Pipeline-1 to outputs
<label @PIPELINE_1>
  # Add User Defined labels to the output record
  <filter **>
    @type record_modifier
    remove_keys _dummy_
    <record>
      _dummy_ ${record['openshift']={"labels"=>{}} unless record['openshift'];record['openshift']['labels'] = {"1"=>"2"} }
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @INFRA_ES
  </match>
</label>
`))
	})

	It("should generate multiple labels for a single pipeline", func() {
		pipelines := []logging.PipelineSpec{
			{
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"infra-es"},
				Labels:     map[string]string{"1": "2", "3": "4"},
				Name:       "Pipeline-1",
			},
		}
		c := PipelineToOutputs(&logging.ClusterLogForwarderSpec{
			Pipelines: pipelines,
		}, nil)
		got, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(got).To(EqualTrimLines(`
# Copying pipeline Pipeline-1 to outputs
<label @PIPELINE_1>
  # Add User Defined labels to the output record
  <filter **>
    @type record_modifier
    remove_keys _dummy_
    <record>
      _dummy_ ${record['openshift']={"labels"=>{}} unless record['openshift'];record['openshift']['labels'] = {"1"=>"2","3"=>"4"} }
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @INFRA_ES
  </match>
</label>
`))
	})

	It("should generate multiple labels for multiple pipelines", func() {
		pipelines := []logging.PipelineSpec{
			{
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"infra-es"},
				Labels:     map[string]string{"1": "2", "3": "4"},
				Name:       "Pipeline-1",
			},
			{
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"infra-es"},
				Labels:     map[string]string{"5": "6", "7": "8"},
				Name:       "Pipeline-2",
			},
		}
		c := PipelineToOutputs(&logging.ClusterLogForwarderSpec{
			Pipelines: pipelines,
		}, nil)
		got, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(got).To(EqualTrimLines(`
# Copying pipeline Pipeline-1 to outputs
<label @PIPELINE_1>
  # Add User Defined labels to the output record
  <filter **>
    @type record_modifier
    remove_keys _dummy_
    <record>
      _dummy_ ${record['openshift']={"labels"=>{}} unless record['openshift'];record['openshift']['labels'] = {"1"=>"2","3"=>"4"} }
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @INFRA_ES
  </match>
</label>

# Copying pipeline Pipeline-2 to outputs
<label @PIPELINE_2>
  # Add User Defined labels to the output record
  <filter **>
    @type record_modifier
    remove_keys _dummy_
    <record>
      _dummy_ ${record['openshift']={"labels"=>{}} unless record['openshift'];record['openshift']['labels'] = {"5"=>"6","7"=>"8"} }
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @INFRA_ES
  </match>
</label>
`))
	})
})

package fluentd

import (
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var _ = Describe("Generating pipeline to output labels", func() {
	var (
		configGenerator *ConfigGenerator
		err             error
	)
	BeforeEach(func() {
		configGenerator, err = NewConfigGenerator(false, false, false)
		Expect(err).To(BeNil())
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
		got, err := configGenerator.generatePipelineToOutputLabels(pipelines)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @PIPELINE_1>
  <match **>
    @type copy

    <store>
      @type relabel
      @label @INFRA_ES
    </store>
  </match>
</label>`))
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
		got, err := configGenerator.generatePipelineToOutputLabels(pipelines)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @PIPELINE_1>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"1":"2"} }
    </record>
  </filter>
  <match **>
    @type copy

    <store>
      @type relabel
      @label @INFRA_ES
    </store>
  </match>
</label>`))
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
		got, err := configGenerator.generatePipelineToOutputLabels(pipelines)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @PIPELINE_1>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"1":"2","3":"4"} }
    </record>
  </filter>
  <match **>
    @type copy

    <store>
      @type relabel
      @label @INFRA_ES
    </store>
  </match>
</label>`))
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
		got, err := configGenerator.generatePipelineToOutputLabels(pipelines)
		Expect(err).To(BeNil())
		Expect(got).To(BeEquivalentTo([]string{`<label @PIPELINE_1>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"1":"2","3":"4"} }
    </record>
  </filter>
  <match **>
    @type copy
    <store>
      @type relabel
      @label @INFRA_ES
    </store>
  </match>
</label>`, `<label @PIPELINE_2>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"5":"6","7":"8"} }
    </record>
  </filter>
  <match **>
    @type copy
    <store>
      @type relabel
      @label @INFRA_ES
    </store>
  </match>
</label>`}))
	})
})

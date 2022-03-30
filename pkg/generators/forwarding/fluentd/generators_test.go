package fluentd

import (
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"k8s.io/apimachinery/pkg/util/sets"

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
var _ = Describe("Generating source to pipeline label", func() {
	var (
		configGenerator *ConfigGenerator
		err             error
	)
	BeforeEach(func() {
		configGenerator, err = NewConfigGenerator(false, false, false)
		Expect(err).To(BeNil())
	})

	It("should generate no namespace filtering routes with a match all route", func() {
		pipelines := logging.RouteMap{
			logging.InputNameApplication: sets.NewString("pipeline-1"),
		}
		nsMap := logging.RouteMap{
			"": sets.NewString("pipeline-1"),
		}
		got, err := configGenerator.generateSourceToPipelineLabels(pipelines, nsMap)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @_APPLICATION>
  <match **>
    @type copy

    <store>
      @type relabel
      @label @PIPELINE_1
    </store>
  </match>
</label>`))
	})
	It("should generate a namespace filtering route without match all routes", func() {
		pipelines := logging.RouteMap{
			logging.InputNameApplication: sets.NewString("pipeline-1"),
		}
		nsMap := logging.RouteMap{
			"project-1": sets.NewString("pipeline-1"),
		}
		got, err := configGenerator.generateSourceToPipelineLabels(pipelines, nsMap)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @_APPLICATION>
    <match **>
      @type copy

      <store>
        @type relabel
        @label @_APPLICATION_NAMESPACE_FILTERING
      </store>
    </match>
  </label>

  <label @_APPLICATION_NAMESPACE_FILTERING>
    <match kubernetes.**_project-1_**>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_1
      </store>
    </match>
</label>`))
	})
	It("should generate a namespace filtering route with a match all route", func() {
		pipelines := logging.RouteMap{
			logging.InputNameApplication: sets.NewString("pipeline-1", "pipeline-2"),
		}
		nsMap := logging.RouteMap{
			"project-1": sets.NewString("pipeline-1"),
			"":          sets.NewString("pipeline-2"),
		}
		got, err := configGenerator.generateSourceToPipelineLabels(pipelines, nsMap)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @_APPLICATION>
    <match **>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_2
      </store>

      <store>
        @type relabel
        @label @_APPLICATION_NAMESPACE_FILTERING
      </store>
    </match>
  </label>

  <label @_APPLICATION_NAMESPACE_FILTERING>
    <match kubernetes.**_project-1_**>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_1
      </store>
    </match>
</label>`))
	})
	It("should generate namespace filtering routes with a match all route", func() {
		pipelines := logging.RouteMap{
			logging.InputNameApplication: sets.NewString("pipeline-1", "pipeline-2"),
		}
		nsMap := logging.RouteMap{
			"project-1": sets.NewString("pipeline-1"),
			"project-2": sets.NewString("pipeline-2"),
			"":          sets.NewString("pipeline-2"),
		}
		got, err := configGenerator.generateSourceToPipelineLabels(pipelines, nsMap)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @_APPLICATION>
    <match **>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_2
      </store>

      <store>
        @type relabel
        @label @_APPLICATION_NAMESPACE_FILTERING
      </store>
    </match>
  </label>

  <label @_APPLICATION_NAMESPACE_FILTERING>
    <match kubernetes.**_project-1_**>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_1
      </store>
    </match>
    <match kubernetes.**_project-2_**>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_2
      </store>
    </match>
</label>`))
	})
	It("should generate namespace filtering routes with match all routes", func() {
		pipelines := logging.RouteMap{
			logging.InputNameApplication: sets.NewString("pipeline-1", "pipeline-2", "pipeline-3"),
		}
		nsMap := logging.RouteMap{
			"project-1": sets.NewString("pipeline-1"),
			"project-2": sets.NewString("pipeline-2"),
			"":          sets.NewString("pipeline-2", "pipeline-3"),
		}
		got, err := configGenerator.generateSourceToPipelineLabels(pipelines, nsMap)
		Expect(err).To(BeNil())
		Expect(got[0]).To(EqualTrimLines(`<label @_APPLICATION>
    <match **>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_2
      </store>
      <store>
        @type relabel
        @label @PIPELINE_3
      </store>

      <store>
        @type relabel
        @label @_APPLICATION_NAMESPACE_FILTERING
      </store>
    </match>
  </label>

  <label @_APPLICATION_NAMESPACE_FILTERING>
    <match kubernetes.**_project-1_**>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_1
      </store>
    </match>
    <match kubernetes.**_project-2_**>
      @type copy
      <store>
        @type relabel
        @label @PIPELINE_2
      </store>
    </match>
</label>`))
	})

})
var _ = Describe("mapAppNamespacesToPipelines", func() {
	var (
		forwarder *logging.ClusterLogForwarder
	)
	Context("with default inputs", func() {

		It("should correctly map application namespaces to pipelines", func() {
			forwarder = &logging.ClusterLogForwarder{
				Spec: logging.ClusterLogForwarderSpec{
					Pipelines: []logging.PipelineSpec{
						{
							Name:      "pipeline-foo",
							InputRefs: []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						},
						{
							Name:      "pipeline-bar",
							InputRefs: []string{logging.InputNameAudit},
						},
					},
				},
			}
			nsMap := logging.RouteMap{
				"": sets.NewString("pipeline-foo"),
			}

			Expect(mapAppNamespacesToPipelines(&forwarder.Spec)).To(BeEquivalentTo(nsMap))
		})
	})
	Context("explicitly defining inputs", func() {

		It("should correctly map application namespaces to pipelines", func() {
			forwarder = &logging.ClusterLogForwarder{
				Spec: logging.ClusterLogForwarderSpec{
					Outputs: []logging.OutputSpec{
						{Name: "output1"},
						{Name: "output2"},
					},
					Inputs: []logging.InputSpec{
						{
							Name: "input1",
							Application: &logging.Application{
								Namespaces: []string{"project-1", "project-2"},
							},
						},
						{
							Name: "input2",
							Application: &logging.Application{
								Namespaces: []string{"project-2", "project-3"},
							},
						},
					},
					Pipelines: []logging.PipelineSpec{
						{
							Name:      "pipeline-foo",
							InputRefs: []string{"input1", logging.InputNameInfrastructure},
						},
						{
							Name:      "pipeline-bar",
							InputRefs: []string{"input2"},
						},
					},
				},
			}
			nsMap := logging.RouteMap{
				"project-1": sets.NewString("pipeline-foo"),
				"project-2": sets.NewString("pipeline-foo", "pipeline-bar"),
				"project-3": sets.NewString("pipeline-bar"),
			}

			Expect(mapAppNamespacesToPipelines(&forwarder.Spec)).To(BeEquivalentTo(nsMap))
		})
	})
})

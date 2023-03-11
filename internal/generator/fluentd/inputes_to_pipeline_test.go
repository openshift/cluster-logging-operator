package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating conf to route entries to pipelines", func() {

	DescribeTable("#SourceTypeToPipeline", func(sourceType string, conf helpers.ConfGenerateTest) {
		helpers.TestGenerateConfWith(func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
			return []generator.Element{SourceTypeToPipeline(sourceType, &clfspec, generator.NoOptions)}
		})(conf)
	},
		Entry("should deep copy when parsing is enabled and more then one pipeline", logging.InputNameApplication, helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{logging.InputNameApplication},
						Name:      "one",
						Parse:     JSONParseType,
					},
					{
						InputRefs: []string{logging.InputNameApplication},
						Name:      "two",
					},
				},
			},
			ExpectedConf: `
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
	copy_mode deep
    <store>
      @type relabel
      @label @ONE
    </store>

    <store>
      @type relabel
      @label @TWO
    </store>
  </match>
</label>`,
		}),
	)
})

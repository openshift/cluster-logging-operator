package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Multiline exception detection", func() {

	var (
		g       generator.Generator
		err     error
		results string
		op      generator.Options
		fwd     logging.ClusterLogForwarderSpec
		e       []generator.Element
	)

	It("should include a pipeline to evaluate container logs", func() {

		e = MultilineDetectExceptions(&fwd, op)
		results, err = g.GenerateConf(e...)
		Expect(err).To(BeNil())
		Expect(results).To(matchers.EqualTrimLines(
			`
<label @_MULITLINE_DETECT>
  <match kubernetes.**>
    @id multiline-detect-except
    @type detect_exceptions
    remove_tag_prefix 'kubernetes'
    message log
    force_line_breaks true
    multiline_flush_interval .2
  </match>
  <match **>
    @type relabel
    @label @INGRESS
  </match>
</label>
`))
	})
})

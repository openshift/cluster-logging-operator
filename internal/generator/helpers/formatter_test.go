package helpers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("FormatFluentConf", func() {

	const unformatted = "\n<system>\n  log_level \"#{ENV['LOG_LEVEL'] || 'warn'}\"\n</system>\n\n" +
		"# Prometheus Monitoring\n<label @ES_1>\n  #remove structured field if present\n  <filter **>\n" +
		"    @type record_modifier\n    remove_keys structured\n  </filter>\n  \n" +
		"  #flatten labels to prevent field explosion in ES\n  <filter **>\n    @type record_transformer\n" +
		"    enable_ruby true\n    <record>\n      foo bar\n    </record>\n  </filter>\n</label>"
	It("should do nothing for an empty string", func() {
		Expect(FormatFluentConf("")).To(BeEmpty())
	})

	It("should format the fluent configuration", func() {
		Expect(FormatFluentConf(unformatted)).To(matchers.EqualDiff(`
<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<label @ES_1>
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>

  #flatten labels to prevent field explosion in ES
  <filter **>
    @type record_transformer
    enable_ruby true
    <record>
      foo bar
    </record>
  </filter>
</label>`))
	})

})

package helpers

import (
	. "github.com/onsi/ginkgo"
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

var _ = Describe("FormatVectorToml", func() {
	const unformatted = `
expire_metrics_secs = 60

data_dir = "/var/lib/vector/testhack-nttnlemu/instance"


[api]
enabled = true

[sources.internal_metrics]
type = "internal_metrics"

# Logs from containers (including openshift containers)
# to this output
[sinks.output_http]
type = "http"
inputs = ["output_http_dedot"]
uri = "http://localhost:8090"
method = "post"


[sinks.output_http.encoding]
codec = "json"



[sinks.output_http.buffer]

when_full = "drop_newest"



[sinks.output_http.request]





headers = {"k1"="v1"}


[transforms.add_nodename_to_metric]
type = "remap"
inputs = ["internal_metrics"]
source = '''
    .tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''
`
	It("should do nothing for an empty string", func() {
		Expect(FormatVectorToml("")).To(BeEmpty())
	})
	It("should format the vector.toml configuration", func() {
		Expect(FormatVectorToml(unformatted)).To(matchers.EqualDiff(`expire_metrics_secs = 60
data_dir = "/var/lib/vector/testhack-nttnlemu/instance"

[api]
enabled = true

[sources.internal_metrics]
type = "internal_metrics"

# Logs from containers (including openshift containers)
# to this output
[sinks.output_http]
type = "http"
inputs = ["output_http_dedot"]
uri = "http://localhost:8090"
method = "post"

[sinks.output_http.encoding]
codec = "json"

[sinks.output_http.buffer]
when_full = "drop_newest"

[sinks.output_http.request]
headers = {"k1"="v1"}

[transforms.add_nodename_to_metric]
type = "remap"
inputs = ["internal_metrics"]
source = '''
    .tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''`))
	})
})

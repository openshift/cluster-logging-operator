package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vectorapi "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/test"
	"sigs.k8s.io/yaml"
)

const configYaml = `
expire_metrics_secs: 60
data_dir: "/var/lib/vector/openshift-logging/collector"
api:
  enabled: true
  address: "127.0.0.1:8686"
log_schema:
  host_key: "hostname"
secret:
  kubernetes_secret:
    type: "file"
    path: "/var/run/ocp-collector/secrets"
sources:
  internal_metrics:
    type: "internal_metrics"
    scrape_interval_secs: 2
transforms:
  bar:
    inputs: ["internal_metrics"]
    type: "remap"
    source: "abc 123"
sinks:
  foo:
    inputs: ["bar"]
    type: "http"
    uri: "http://nowhwere:123"
`

var _ = Describe("Config", func() {

	It("should marshal a programmatically built config to the expected YAML", func() {
		config := vectorapi.NewConfig(func(c *vectorapi.Config) {
			c.ExpireMetricsSec = 60
			c.DataDir = "/var/lib/vector/openshift-logging/collector"
			c.Api = &vectorapi.Api{
				Enabled: true,
				Address: "127.0.0.1:8686",
			}
			c.LogSchema = &vectorapi.LogSchema{
				HostKey: "hostname",
			}
			c.Secret = map[string]*vectorapi.Secret{
				"kubernetes_secret": {
					Type: "file",
					Path: "/var/run/ocp-collector/secrets",
				},
			}
			c.Sources = vectorapi.Sources{
				"internal_metrics": &sources.InternalMetrics{
					Type:               types.SourceTypeInternalMetrics,
					ScrapeIntervalSecs: 2,
				},
			}
			c.Transforms = vectorapi.Transforms{
				"bar": transforms.NewRemap("abc 123", "internal_metrics"),
			}
			c.Sinks = vectorapi.Sinks{
				"foo": sinks.NewHttp("http://nowhwere:123", nil, "bar"),
			}
		})
		Expect(test.YAMLString(config)).To(MatchYAML(configYaml))
	})

	It("marshals a VRL source containing TOML literal delimiters and newlines without breakout", func() {
		vrl := "if true {\n  .msg = '''not a toml literal'''\n}"
		config := vectorapi.NewConfig(func(c *vectorapi.Config) {
			c.Transforms = vectorapi.Transforms{
				"bar": &transforms.Remap{
					Type:   types.TransformTypeRemap,
					Inputs: []string{"internal_metrics"},
					Source: transforms.VrlString(vrl),
				},
			}
		})
		b, err := yaml.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		var generic map[string]interface{}
		Expect(yaml.Unmarshal(b, &generic)).To(Succeed())

		// No structural breakout: the only top-level key is the transforms map
		Expect(generic).To(HaveLen(1))
		Expect(generic).To(HaveKey("transforms"))

		// The VRL source parses back to the identical string, delimiters intact
		bar := generic["transforms"].(map[string]interface{})["bar"].(map[string]interface{})
		Expect(bar).To(HaveKeyWithValue("source", vrl))
	})

})

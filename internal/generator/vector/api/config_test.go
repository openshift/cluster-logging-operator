package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vectorapi "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
	"github.com/openshift/cluster-logging-operator/test"
)

const (
	configToml = `
expire_metrics_secs = 60
data_dir = "/var/lib/vector/openshift-logging/collector"

[api]
enabled = true

# Load sensitive data from files
[secret.kubernetes_secret]
type = "file"
path = "/var/run/ocp-collector/secrets"

[sources.internal_metrics]
type = "internal_metrics"
scrape_interval_seconds = 2

[transforms.bar]
inputs = ["internal_metrics"]
type = "remap"
source = "abc 123"

[sinks.foo]
inputs = ["bar"]
type = "http"
uri = "http://nowhwere:123"
`
	configYaml = `
expire_metrics_secs: 60
data_dir: "/var/lib/vector/openshift-logging/collector"
api:
  enabled: true
secret:
  kubernetes_secret:
    type: "file"
    path: "/var/run/ocp-collector/secrets"
sources:
  internal_metrics:
    type: "internal_metrics"
    scrape_interval_seconds: 2
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
)

var _ = Describe("Config", func() {

	It("should high level round-trip serialization without field loss", func() {
		config := &vectorapi.Config{}
		toml.MustUnmarshal(configToml, config)
		Expect(test.YAMLString(config)).To(MatchYAML(configYaml))
	})

})

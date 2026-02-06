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
base_path = "/var/run/ocp-collector/secrets"

[sources.internal_metrics]
type = "internal_metrics"
scrap_interval_seconds = 2

[transforms.bar]
type = "remap"

[sinks.foo]
type = "foo"
`
	configYaml = `
expire_metrics_secs: 60
data_dir: "/var/lib/vector/openshift-logging/collector"
api:
  enabled: true
secret:
  kubernetes_secret:
    type: "file"
    base_path: "/var/run/ocp-collector/secrets"
sources:
  internal_metrics:
    type: "internal_metrics"
    scrap_interval_seconds: 2
transforms:
  bar:
    type: "remap"
sinks:
  foo:
    type: "foo"
`
)

var _ = Describe("Config", func() {

	It("should highlevel roundtrip serialization without field loss", func() {
		config := &vectorapi.Config{}
		toml.MustUnMarshal(configToml, config)
		Expect(test.YAMLString(config)).To(MatchYAML(configYaml))
	})

})

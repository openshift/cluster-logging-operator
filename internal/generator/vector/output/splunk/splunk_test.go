package splunk

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

// #nosec G101
const hecToken = "VS0BNth3wCGF0eol0MuK07SHIrhYwCPHFWMG"

var _ = Describe("Generating vector config for Splunk output", func() {

	const (
		splunkSink = `
[sinks.splunk_hec]
type = "splunk_hec"
inputs = ["pipelineName"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"
[sinks.splunk_hec.encoding]
codec = "json"
`
		splunkSinkTls = `
[sinks.splunk_hec]
type = "splunk_hec"
inputs = ["pipelineName"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"
[sinks.splunk_hec.encoding]
codec = "json"
[sinks.splunk_hec.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/ca-bundle.crt"
`
		splunkSinkPassphrase = `
[sinks.splunk_hec]
type = "splunk_hec"
inputs = ["pipelineName"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"

[sinks.splunk_hec.encoding]
codec = "json"

[sinks.splunk_hec.tls]
enabled = true
key_pass = "junk"
`
	)

	var (
		g generator.Generator

		output = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSplunk,
			Name: "splunk_hec",
			URL:  "https://splunk-web:8088/endpoint",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Splunk: &loggingv1.Splunk{},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-splunk-secret",
			},
		}

		outputWithPassphrase = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSplunk,
			Name: "splunk_hec",
			URL:  "https://splunk-web:8088/endpoint",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Splunk: &loggingv1.Splunk{},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-splunk-secret-passphrase",
			},
		}
		outputWithTls = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSplunk,
			Name: "splunk_hec",
			URL:  "https://splunk-web:8088/endpoint",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Splunk: &loggingv1.Splunk{},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-splunk-secret-tls",
			},
		}

		secrets = map[string]*corev1.Secret{
			output.Secret.Name: {
				Data: map[string][]byte{
					"hecToken": []byte(hecToken),
				},
			},
			outputWithTls.Secret.Name: {
				Data: map[string][]byte{
					"hecToken":      []byte(hecToken),
					"tls.key":       []byte("junk"),
					"tls.crt":       []byte("junk"),
					"ca-bundle.crt": []byte("junk"),
				},
			},
			outputWithPassphrase.Secret.Name: {
				Data: map[string][]byte{
					"hecToken":   []byte(hecToken),
					"passphrase": []byte("junk"),
				},
			},
		}
	)

	Context("splunk config", func() {
		BeforeEach(func() {
			g = generator.MakeGenerator()
		})

		It("should provide a valid config", func() {
			element := Conf(output, []string{"pipelineName"}, secrets[output.Secret.Name], nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSink))
		})

		It("should provide a valid config with passphrase", func() {
			element := Conf(outputWithPassphrase, []string{"pipelineName"}, secrets[outputWithPassphrase.Secret.Name], nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkPassphrase))
		})

		It("should provide a valid config with TLS", func() {
			element := Conf(outputWithTls, []string{"pipelineName"}, secrets[outputWithTls.Secret.Name], nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkTls))
		})
	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector for Splunk Conf Generation")
}

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
		splunkDedot = `
[transforms.splunk_hec_dedot]
type = "lua"
inputs = ["pipelineName"]
version = "2"
hooks.init = "init"
hooks.process = "process"
source = '''
    function init()
        count = 0
    end
    function process(event, emit)
        count = count + 1
        event.log.openshift.sequence = count
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
		dedot(event.log.kubernetes.namespace_labels)
        dedot(event.log.kubernetes.labels)
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "[./]", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
'''
`
		splunkSink = splunkDedot + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_dedot"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"
[sinks.splunk_hec.encoding]
codec = "json"
`
		splunkSinkTls = splunkDedot + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_dedot"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"
[sinks.splunk_hec.encoding]
codec = "json"
[sinks.splunk_hec.tls]
key_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/ca-bundle.crt"
`
		splunkSinkTlsSkipVerify = splunkDedot + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_dedot"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"
[sinks.splunk_hec.encoding]
codec = "json"
[sinks.splunk_hec.tls]
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/ca-bundle.crt"
`
		splunkSinkTlsSkipVerifyNoCert = splunkDedot + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_dedot"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = ""
[sinks.splunk_hec.encoding]
codec = "json"
[sinks.splunk_hec.tls]
verify_certificate = false
verify_hostname = false
`
		splunkSinkPassphrase = splunkDedot + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_dedot"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + hecToken + `"

[sinks.splunk_hec.encoding]
codec = "json"

[sinks.splunk_hec.tls]
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

		outputWithTlsSkipVerify = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSplunk,
			Name: "splunk_hec",
			URL:  "https://splunk-web:8088/endpoint",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Splunk: &loggingv1.Splunk{},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-splunk-secret-tls",
			},
			TLS: &loggingv1.OutputTLSSpec{
				InsecureSkipVerify: true,
			},
		}

		outputWithTlsSkipVerifyNoCert = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSplunk,
			Name: "splunk_hec",
			URL:  "https://splunk-web:8088/endpoint",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Splunk: &loggingv1.Splunk{},
			},
			//Secret: &loggingv1.OutputSecretSpec{
			//	Name: "vector-splunk-secret",
			//},
			TLS: &loggingv1.OutputTLSSpec{
				InsecureSkipVerify: true,
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

		It("should provide a valid config with tls.insecureSkipVerify=true", func() {
			element := Conf(outputWithTlsSkipVerify, []string{"pipelineName"}, secrets[outputWithTls.Secret.Name], nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkTlsSkipVerify))
		})

		It("should provide a valid config with tls.insecureSkipVerify=true without secret", func() {
			element := Conf(outputWithTlsSkipVerifyNoCert, []string{"pipelineName"}, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkTlsSkipVerifyNoCert))
		})
	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector for Splunk Conf Generation")
}

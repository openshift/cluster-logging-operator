package splunk

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

const (
	// #nosec G101
	expHecToken = "VS0BNth3wCGF0eol0MuK07SHIrhY$$wCPHFWMG"
	// #nosec G101
	hecToken = "VS0BNth3wCGF0eol0MuK07SHIrhY$wCPHFWMG"
)

var _ = Describe("Generating vector config for Splunk output", func() {

	const (
		splunkDedot = `
[transforms.splunk_hec_dedot]
type = "remap"
inputs = ["pipelineName"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''
`
		fixTimestamp = `
	# Ensure timestamp field well formatted for Splunk
	[transforms.splunk_hec_timestamp]
	type = "remap"
	inputs = ["splunk_hec_dedot"]
	source = '''

	ts, err = parse_timestamp(.@timestamp,"%+")
	if err != null {
		log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
	} else {
		.@timestamp = ts
	}

	'''
`
		splunkSink = splunkDedot + fixTimestamp + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_timestamp"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + expHecToken + `"
timestamp_key = "@timestamp"
[sinks.splunk_hec.encoding]
codec = "json"
`
		splunkSinkTls = splunkDedot + fixTimestamp + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_timestamp"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + expHecToken + `"
timestamp_key = "@timestamp"
[sinks.splunk_hec.encoding]
codec = "json"

[sinks.splunk_hec.tls]
key_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/ca-bundle.crt"
`
		splunkSinkTlsSkipVerify = splunkDedot + fixTimestamp + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_timestamp"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + expHecToken + `"
timestamp_key = "@timestamp"
[sinks.splunk_hec.encoding]
codec = "json"

[sinks.splunk_hec.tls]
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-splunk-secret-tls/ca-bundle.crt"
`
		splunkSinkTlsSkipVerifyNoCert = splunkDedot + fixTimestamp + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_timestamp"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = ""
timestamp_key = "@timestamp"
[sinks.splunk_hec.encoding]
codec = "json"

[sinks.splunk_hec.tls]
verify_certificate = false
verify_hostname = false
`
		splunkSinkPassphrase = splunkDedot + fixTimestamp + `
[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_timestamp"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + expHecToken + `"
timestamp_key = "@timestamp"

[sinks.splunk_hec.encoding]
codec = "json"

[sinks.splunk_hec.tls]
key_pass = "junk"
`
	)

	var (
		g framework.Generator

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
			g = framework.MakeGenerator()
		})

		It("should provide a valid config", func() {
			element := New(vectorhelpers.FormatComponentID(output.Name), output, []string{"pipelineName"}, secrets[output.Secret.Name], nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSink))
		})

		It("should provide a valid config with passphrase", func() {
			element := New(vectorhelpers.FormatComponentID(output.Name), outputWithPassphrase, []string{"pipelineName"}, secrets[outputWithPassphrase.Secret.Name], nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkPassphrase))
		})

		It("should provide a valid config with TLS", func() {
			element := New(vectorhelpers.FormatComponentID(output.Name), outputWithTls, []string{"pipelineName"}, secrets[outputWithTls.Secret.Name], nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkTls))
		})

		It("should provide a valid config with tls.insecureSkipVerify=true", func() {
			element := New(vectorhelpers.FormatComponentID(output.Name), outputWithTlsSkipVerify, []string{"pipelineName"}, secrets[outputWithTls.Secret.Name], nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkTlsSkipVerify))
		})

		It("should provide a valid config with tls.insecureSkipVerify=true without secret", func() {
			element := New(vectorhelpers.FormatComponentID(output.Name), outputWithTlsSkipVerifyNoCert, []string{"pipelineName"}, nil, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(splunkSinkTlsSkipVerifyNoCert))
		})

		Context("with custom index", func() {
			var (
				splunkOutputSpec loggingv1.OutputSpec
				splunkIndexRemap = `
# Set Splunk Index
[transforms.splunk_hec_add_splunk_index]
type = "remap"
inputs = ["pipelineName"]			
`
				splunkWithIndexDedot = `

[transforms.splunk_hec_dedot]
type = "remap"
inputs = ["splunk_hec_add_splunk_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''
# Ensure timestamp field well formatted for Splunk
[transforms.splunk_hec_timestamp]
type = "remap"
inputs = ["splunk_hec_dedot"]
source = '''

ts, err = parse_timestamp(.@timestamp,"%+")
if err != null {
	log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
} else {
	.@timestamp = ts
}

'''

[sinks.splunk_hec]
type = "splunk_hec_logs"
inputs = ["splunk_hec_timestamp"]
endpoint = "https://splunk-web:8088/endpoint"
compression = "none"
default_token = "` + expHecToken + `"
index = "{{ write_index }}"
timestamp_key = "@timestamp"
[sinks.splunk_hec.encoding]
codec = "json"
except_fields = ["write_index"]
`

				splunkSinkIndexKey = `
source = '''
val = .kubernetes.namespace_name
if !is_null(val) {
	.write_index = val
} else {
	.write_index = ""
}
'''
`
				splunkSinkIndexName = `
source = '''
	.write_index = "custom-index"
'''
`
			)

			BeforeEach(func() {
				splunkOutputSpec = loggingv1.OutputSpec{
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
			})

			It("should provide a valid config with indexKey specified", func() {
				splunkOutputSpec.Splunk.IndexKey = "kubernetes.namespace_name"
				element := New(vectorhelpers.FormatComponentID(output.Name), splunkOutputSpec, []string{"pipelineName"}, secrets[output.Secret.Name], nil, nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(splunkIndexRemap + splunkSinkIndexKey + splunkWithIndexDedot))
			})

			It("should provide a valid config with indexName specified", func() {
				splunkOutputSpec.Splunk.IndexName = "custom-index"
				element := New(vectorhelpers.FormatComponentID(output.Name), splunkOutputSpec, []string{"pipelineName"}, secrets[output.Secret.Name], nil, nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(splunkIndexRemap + splunkSinkIndexName + splunkWithIndexDedot))
			})
		})
	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector for Splunk Conf Generation")
}

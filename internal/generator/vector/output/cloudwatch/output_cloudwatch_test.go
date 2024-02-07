package cloudwatch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tls "github.com/openshift/cluster-logging-operator/internal/tls"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	keyId           = "AKIAIOSFODNN7EXAMPLE"
	keySecret       = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
	defaultTLS      = "VersionTLS12"
	defaultCiphers  = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
	vectorTLSSecret = "vector-cw-secret-tls"
	transformBegin  = `
# Cloudwatch Group and Stream Names
[transforms.cw_normalize_group_and_streams]
type = "remap"
inputs = ["cw-forward"]
source = '''
  .group_name = "default"
  .stream_name = "default"

  if (.file != null) {
   .file = "kubernetes" + replace!(.file, "/", ".")
   .stream_name = del(.file)
  }
`
	transformEnd = `
  if ( .tag == ".journal.system" ) {
   .stream_name =  ( .hostname + .tag ) ?? .stream_name
  }
  del(.tag)
  del(.source_type)
'''
`

	dedotted = `
[transforms.cw_dedot]
type = "remap"
inputs = ["cw_normalize_group_and_streams"]
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
	cwSinkKeyId = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
healthcheck.enabled = false
`
	cwBufferAndRequest = `
[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
`
	cwSinkRole = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
`
	cwSinkKeyIdTLS = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2

[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
key_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/ca-bundle.crt"
`
	cwSinkKeyIdTLSNoCerts = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
`

	cwSinkKeyIdTLSInsecure = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/ca-bundle.crt"
`
	cwSinkRoleTLS = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
key_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/ca-bundle.crt"
`
	cwSinkRoleTLSInsecure = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/ca-bundle.crt"
`
	cwSinkRoleTLSCredentials = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
key_file = "/var/run/ocp-collector/secrets/vector-tls-credentials/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-tls-credentials/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-tls-credentials/ca-bundle.crt"
`
	cwSinkRoleTLSInsecureCredentials = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_dedot"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
healthcheck.enabled = false

[sinks.cw.buffer]
when_full = "drop_newest"

[sinks.cw.request]
retry_attempts = 17
concurrency = 2
[sinks.cw.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/vector-tls-credentials/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-tls-credentials/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-tls-credentials/ca-bundle.crt"
`
)

var _ = Describe("Generating vector config for cloudwatch output", func() {
	var (
		g  framework.Generator
		op = framework.Options{
			framework.MinTLSVersion: string(tls.DefaultMinTLSVersion),
			framework.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
		}
		secrets      map[string]*corev1.Secret
		groupPrefix  = "all-logs"
		pipelineName = []string{"cw-forward"}

		output = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeCloudwatch,
			Name: "cw",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Cloudwatch: &loggingv1.Cloudwatch{
					Region:      "us-east-test",
					GroupPrefix: &groupPrefix,
					GroupBy:     loggingv1.LogGroupByLogType,
				},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-cw-secret",
			},
		}
	)

	BeforeEach(func() {
		secrets = map[string]*corev1.Secret{
			output.Secret.Name: {
				Data: map[string][]byte{
					"aws_access_key_id":     []byte(keyId),
					"aws_secret_access_key": []byte(keySecret),
				},
			},
			vectorTLSSecret: {
				Data: map[string][]byte{
					"aws_access_key_id":     []byte(keyId),
					"aws_secret_access_key": []byte(keySecret),
					"tls.crt":               []byte("-- crt-- "),
					"tls.key":               []byte("-- key-- "),
					"ca-bundle.crt":         []byte("-- ca-bundle -- "),
				},
			},
		}
	})

	Context("with a group prefix", func() {
		BeforeEach(func() {
			g = framework.MakeGenerator()

		})

		Context("grouped by log type", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
			})

			It("should provide a valid config", func() {
				expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "` + groupPrefix + `.audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "` + groupPrefix + `.infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkKeyId + `
` + cwBufferAndRequest + `
`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})

		Context("grouped by namespace name", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByNamespaceName
			})

			It("should provide a valid config", func() {
				expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "` + groupPrefix + `." + .kubernetes.namespace_name ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "` + groupPrefix + `.audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "` + groupPrefix + `.infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkKeyId + `
` + cwBufferAndRequest + `
`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})

		Context("grouped by namespace uuid", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByNamespaceUUID
			})

			It("should provide a valid config", func() {
				expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "` + groupPrefix + `." + .kubernetes.namespace_id ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "` + groupPrefix + `.audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "` + groupPrefix + `.infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkKeyId + `
` + cwBufferAndRequest + `
`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})
	})

	Context("without specifying a prefix", func() {
		BeforeEach(func() {
			g = framework.MakeGenerator()
		})

		Context("grouped by log type without prefix", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
				output.Cloudwatch.GroupPrefix = nil
			})

			It("should provide a valid config", func() {
				expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "" + .log_type ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkKeyId + `
` + cwBufferAndRequest + `
`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})
	})

	Context("using endpoint config", func() {
		endpoint := "https://a-test-endpoint:9200"
		BeforeEach(func() {
			output.URL = endpoint
		})

		It("should provide a valid config", func() {
			expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "" + .log_type ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkKeyId + `
endpoint = "` + endpoint + `"` + cwBufferAndRequest

			element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(expConf))
		})
	})

	Context("with TLS config with default minTLSVersion & ciphers", func() {

		BeforeEach(func() {
			output.Secret.Name = vectorTLSSecret
			output.Cloudwatch.GroupPrefix = nil
			output.URL = ""
		})

		It("should generate an output.tls config block", func() {
			output.Cloudwatch.GroupBy = ""
			secrets[vectorTLSSecret].Data = map[string][]byte{
				"aws_access_key_id":     []byte(keyId),
				"aws_secret_access_key": []byte(keySecret),
			}
			expConf := `
	` + transformBegin + `

	  if ( .log_type == "application" ) {
	   .group_name = ( "" + .log_type ) ?? "application"
	  }
	  if ( .log_type == "audit" ) {
	   .group_name = "audit"
	   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
	  }
	  if ( .log_type == "infrastructure" ) {
	   .group_name = "infrastructure"
	   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
	  }

	` + transformEnd + `

	` + dedotted + `

	` + cwSinkKeyIdTLSNoCerts + `
	`
			element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(expConf))
		})

		It("InsecureSkipVerify omitted", func() {
			expConf := `
	` + transformBegin + `

	  if ( .log_type == "application" ) {
	   .group_name = ( "" + .log_type ) ?? "application"
	  }
	  if ( .log_type == "audit" ) {
	   .group_name = "audit"
	   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
	  }
	  if ( .log_type == "infrastructure" ) {
	   .group_name = "infrastructure"
	   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
	  }

	` + transformEnd + `

	` + dedotted + `

	` + cwSinkKeyIdTLS + `
	`
			element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(expConf))
		})

		It("InsecureSkipVerify set to true", func() {
			output.TLS = &loggingv1.OutputTLSSpec{
				InsecureSkipVerify: true,
			}

			expConf := `
	` + transformBegin + `

	  if ( .log_type == "application" ) {
	   .group_name = ( "" + .log_type ) ?? "application"
	  }
	  if ( .log_type == "audit" ) {
	   .group_name = "audit"
	   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
	  }
	  if ( .log_type == "infrastructure" ) {
	   .group_name = "infrastructure"
	   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
	  }

	` + transformEnd + `

	` + dedotted + `

	` + cwSinkKeyIdTLSInsecure + `
	`
			element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(expConf))
		})

	})
})

var _ = Describe("Generating vector config for cloudwatch sts", func() {
	var (
		g  framework.Generator
		op = framework.Options{
			framework.MinTLSVersion: string(tls.DefaultMinTLSVersion),
			framework.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
		}

		groupPrefix  = "all-logs"
		pipelineName = []string{"cw-forward"}

		output = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeCloudwatch,
			Name: "cw",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Cloudwatch: &loggingv1.Cloudwatch{
					Region:      "us-east-test",
					GroupPrefix: &groupPrefix,
					GroupBy:     loggingv1.LogGroupByLogType,
				},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-cw-secret",
			},
		}

		roleArn = "arn:aws:iam::123456789012:role/my-role-to-assume"
		secrets = map[string]*corev1.Secret{
			output.Secret.Name: {
				Data: map[string][]byte{
					"role_arn": []byte(roleArn),
				},
			},
			vectorTLSSecret: {
				Data: map[string][]byte{
					"role_arn":      []byte(roleArn),
					"tls.crt":       []byte("-- crt-- "),
					"tls.key":       []byte("-- key-- "),
					"ca-bundle.crt": []byte("-- ca-bundle -- "),
				},
			},
		}
	)

	Context("with a role_arn key", func() {
		BeforeEach(func() {
			g = framework.MakeGenerator()
		})
		Context("grouped by log type", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
			})

			It("should provide a valid config", func() {
				expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "` + groupPrefix + `.audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "` + groupPrefix + `.infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkRole + `
`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})

		Context("with TLS config with default minTLSVersion & ciphers", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
				output.Secret.Name = vectorTLSSecret
			})

			It("InsecureSkipVerify false", func() {
				output.TLS = &loggingv1.OutputTLSSpec{
					InsecureSkipVerify: false,
				}
				expConf := `
			` + transformBegin + `

			if ( .log_type == "application" ) {
			.group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
			}
			if ( .log_type == "audit" ) {
			.group_name = "` + groupPrefix + `.audit"
			.stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
			}
			if ( .log_type == "infrastructure" ) {
			.group_name = "` + groupPrefix + `.infrastructure"
			.stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
			}

			` + transformEnd + `

			` + dedotted + `

			` + cwSinkRoleTLS + `
			`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})

			It("InsecureSkipVerify set to false", func() {
				output.TLS = &loggingv1.OutputTLSSpec{
					InsecureSkipVerify: true,
				}
				expConf := `
			` + transformBegin + `

			if ( .log_type == "application" ) {
			.group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
			}
			if ( .log_type == "audit" ) {
			.group_name = "` + groupPrefix + `.audit"
			.stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
			}
			if ( .log_type == "infrastructure" ) {
			.group_name = "` + groupPrefix + `.infrastructure"
			.stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
			}

			` + transformEnd + `

			` + dedotted + `

			` + cwSinkRoleTLSInsecure + `
			`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})

		})
	})

	Context("with credentials key", func() {
		BeforeEach(func() {
			credentialsString := "[default]\nrole_arn = " + roleArn + "\nweb_identity_token_file = /var/run/secrets/token"
			secrets["my-secret"] = &corev1.Secret{
				Data: map[string][]byte{
					"credentials": []byte(credentialsString),
				},
			}
		})
		Context("grouped by log type", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
				output.Secret.Name = "my-secret"
				output.TLS = &loggingv1.OutputTLSSpec{
					InsecureSkipVerify: false,
				}
			})

			It("should provide a valid config", func() {
				expConf := `
` + transformBegin + `

  if ( .log_type == "application" ) {
   .group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
  }
  if ( .log_type == "audit" ) {
   .group_name = "` + groupPrefix + `.audit"
   .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
  }
  if ( .log_type == "infrastructure" ) {
   .group_name = "` + groupPrefix + `.infrastructure"
   .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
  }

` + transformEnd + `

` + dedotted + `

` + cwSinkRole + `
`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})

		Context("with TLS config with default minTLSVersion & ciphers", func() {
			credentialsString := "[default]\nrole_arn = " + roleArn + "\nweb_identity_token_file = /var/run/secrets/token"
			secrets["vector-tls-credentials"] = &corev1.Secret{
				Data: map[string][]byte{
					"credentials":   []byte(credentialsString),
					"tls.crt":       []byte("-- crt-- "),
					"tls.key":       []byte("-- key-- "),
					"ca-bundle.crt": []byte("-- ca-bundle -- "),
				},
			}
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
				output.Secret.Name = "vector-tls-credentials"
			})

			It("InsecureSkipVerify false", func() {
				output.TLS = &loggingv1.OutputTLSSpec{
					InsecureSkipVerify: false,
				}
				expConf := `
			` + transformBegin + `

			if ( .log_type == "application" ) {
			.group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
			}
			if ( .log_type == "audit" ) {
			.group_name = "` + groupPrefix + `.audit"
			.stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
			}
			if ( .log_type == "infrastructure" ) {
			.group_name = "` + groupPrefix + `.infrastructure"
			.stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
			}

			` + transformEnd + `

			` + dedotted + `

			` + cwSinkRoleTLSCredentials + `
			`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})

			It("InsecureSkipVerify set to true", func() {
				output.TLS = &loggingv1.OutputTLSSpec{
					InsecureSkipVerify: true,
				}
				expConf := `
			` + transformBegin + `

			if ( .log_type == "application" ) {
			.group_name = ( "` + groupPrefix + `." + .log_type ) ?? "application"
			}
			if ( .log_type == "audit" ) {
			.group_name = "` + groupPrefix + `.audit"
			.stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
			}
			if ( .log_type == "infrastructure" ) {
			.group_name = "` + groupPrefix + `.infrastructure"
			.stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
			}

			` + transformEnd + `

			` + dedotted + `

			` + cwSinkRoleTLSInsecureCredentials + `
			`
				element := New(helpers.FormatComponentID(output.Name), output, pipelineName, secrets[output.Secret.Name], op)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})

		})
	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}

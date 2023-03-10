package cloudwatch

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tls "github.com/openshift/cluster-logging-operator/internal/tls"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
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
	cwSinkKeyId = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
`
	cwSinkRole = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
`
	cwSinkKeyIdTLS = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
[sinks.cw.tls]
enabled = true
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
key_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/vector-cw-secret-tls/ca-bundle.crt"
`

	cwSinkKeyIdTLSInsecure = `
# Cloudwatch Logs
[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = "` + keyId + `"
auth.secret_access_key = "` + keySecret + `"
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
[sinks.cw.tls]
enabled = true
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
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
[sinks.cw.tls]
enabled = true
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
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
[sinks.cw.tls]
enabled = true
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
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
[sinks.cw.tls]
enabled = true
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
inputs = ["cw_normalize_group_and_streams"]
region = "us-east-test"
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
# role_arn and identity token set via env vars
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false
[sinks.cw.tls]
enabled = true
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
		g  generator.Generator
		op = generator.Options{
			generator.MinTLSVersion: string(tls.DefaultMinTLSVersion),
			generator.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
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
	)

	Context("with a group prefix", func() {
		BeforeEach(func() {
			g = generator.MakeGenerator()

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

` + cwSinkKeyId + `
`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
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

` + cwSinkKeyId + `
`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
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
   .group_name = ( "` + groupPrefix + `." + .kubernetes.namespace_uid ) ?? "application"
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

` + cwSinkKeyId + `
`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
				results, err := g.GenerateConf(element...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(expConf))
			})
		})
	})

	Context("without specifying a prefix", func() {
		BeforeEach(func() {
			g = generator.MakeGenerator()
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

` + cwSinkKeyId + `
`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
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

` + cwSinkKeyId + `
endpoint = "` + endpoint + `"
tls.verify_certificate = false
`
			element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
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

	` + cwSinkKeyIdTLS + `
	`
			element := Conf(output, pipelineName, secrets[output.Secret.Name], op)
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

	` + cwSinkKeyIdTLSInsecure + `
	`
			element := Conf(output, pipelineName, secrets[output.Secret.Name], op)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(expConf))
		})

	})
})

var _ = Describe("Generating vector config for cloudwatch sts", func() {
	var (
		g  generator.Generator
		op = generator.Options{
			generator.MinTLSVersion: string(tls.DefaultMinTLSVersion),
			generator.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
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
			g = generator.MakeGenerator()
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

` + cwSinkRole + `
`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
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
			
			` + cwSinkRoleTLS + `
			`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], op)
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
			
			` + cwSinkRoleTLSInsecure + `
			`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], op)
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

` + cwSinkRole + `
`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], nil)
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
			
			` + cwSinkRoleTLSCredentials + `
			`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], op)
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
			
			` + cwSinkRoleTLSInsecureCredentials + `
			`
				element := Conf(output, pipelineName, secrets[output.Secret.Name], op)
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

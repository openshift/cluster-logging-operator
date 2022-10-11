package cloudwatch

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	keyId          = "AKIAIOSFODNN7EXAMPLE"
	keySecret      = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
	transformBegin = `
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
)

var _ = Describe("Generating vector config for cloudwatch output", func() {
	var (
		g generator.Generator

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
})

var _ = Describe("Generating vector config for cloudwatch sts", func() {
	var (
		g generator.Generator

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
	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}

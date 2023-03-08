package gcl

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	tls "github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	inputPipeline := []string{"application"}
	defaultTLS := "VersionTLS12"
	defaultCiphers := "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		e := []generator.Element{}
		for _, o := range clfspec.Outputs {
			e = generator.MergeElements(e, Conf(o, inputPipeline, secrets[o.Name], op))
		}
		return e
	}
	DescribeTable("For GoogleCloudLogging output", helpers.TestGenerateConfWith(f),
		Entry("with service account token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeGoogleCloudLogging,
						Name: "gcl-1",
						OutputTypeSpec: logging.OutputTypeSpec{
							GoogleCloudLogging: &logging.GoogleCloudLogging{
								BillingAccountID: "billing-1",
								LogID:            "vector-1",
							},
						},
						Secret: &logging.OutputSecretSpec{
							Name: "junk",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"gcl-1": {
					Data: map[string][]byte{
						GoogleApplicationCredentialsKey: []byte("dummy-credentials"),
					},
				},
			},
			ExpectedConf: `
[sinks.gcl_1]
type = "gcp_stackdriver_logs"
inputs = ["application"]
billing_account_id = "billing-1"
credentials_path = "/var/run/ocp-collector/secrets/junk/google-application-credentials.json"
log_id = "vector-1"
severity_key = "level"


[sinks.gcl_1.resource]
type = "k8s_node"
node_name = "{{hostname}}"
`,
		}),
		Entry("with TLS config with default minTLSVersion & ciphers", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeGoogleCloudLogging,
						Name: "gcl-tls",
						OutputTypeSpec: logging.OutputTypeSpec{
							GoogleCloudLogging: &logging.GoogleCloudLogging{
								BillingAccountID: "billing-1",
								LogID:            "vector-1",
							},
						},
						Secret: &logging.OutputSecretSpec{
							Name: "junk",
						},
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"gcl-tls": {
					Data: map[string][]byte{
						GoogleApplicationCredentialsKey: []byte("dummy-credentials"),
						"tls.crt":                       []byte("-- crt-- "),
						"tls.key":                       []byte("-- key-- "),
						"ca-bundle.crt":                 []byte("-- ca-bundle -- "),
					},
				},
			},
			Options: generator.Options{
				generator.MinTLSVersion: string(tls.DefaultMinTLSVersion),
				generator.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
			},
			ExpectedConf: `
[sinks.gcl_tls]
type = "gcp_stackdriver_logs"
inputs = ["application"]
billing_account_id = "billing-1"
credentials_path = "/var/run/ocp-collector/secrets/junk/google-application-credentials.json"
log_id = "vector-1"
severity_key = "level"


[sinks.gcl_tls.resource]
type = "k8s_node"
node_name = "{{hostname}}"

[sinks.gcl_tls.tls]
enabled = true
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/junk/tls.key"
crt_file = "/var/run/ocp-collector/secrets/junk/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/junk/ca-bundle.crt"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}

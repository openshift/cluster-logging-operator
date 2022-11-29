package loki

import (
	"sort"
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("outputLabelConf", func() {
	var (
		loki *logging.Loki
	)
	BeforeEach(func() {
		loki = &logging.Loki{}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(defaultLabelKeys, requiredLabelKeys...)
				sort.Strings(exp)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})
		Context("are spec'd", func() {
			It("should use the ones provided and add the required ones", func() {
				loki.LabelKeys = []string{"foo"}
				exp := append(loki.LabelKeys, requiredLabelKeys...)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})

	})
})

var _ = Describe("Generate vector config", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], generator.NoOptions)
	}
	DescribeTable("for Loki output", helpers.TestGenerateConfWith(f),
		Entry("with default labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: `
[transforms.loki_receiver_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_remap"]
endpoint = "https://logs-us-west1.grafana.net"
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.loki_receiver.encoding]
codec = "json"

[sinks.loki_receiver.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

# Basic Auth Config
[sinks.loki_receiver.auth]
strategy = "basic"
user = "username"
password = "password"
`,
		}),
		Entry("with custom labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-receiver",
						},
						OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
							LabelKeys: []string{"kubernetes.labels.app", "kubernetes.container_name"},
						}},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: `
[transforms.loki_receiver_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_remap"]
endpoint = "https://logs-us-west1.grafana.net"
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.loki_receiver.encoding]
codec = "json"

[sinks.loki_receiver.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_labels_app = "{{kubernetes.labels.app}}"

# Basic Auth Config
[sinks.loki_receiver.auth]
strategy = "basic"
user = "username"
password = "password"
`,
		}),
		Entry("with tenant id", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-receiver",
						},
						OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
							TenantKey: "foo.bar.baz",
						}},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: `
[transforms.loki_receiver_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_remap"]
endpoint = "https://logs-us-west1.grafana.net"
out_of_order_action = "accept"
healthcheck.enabled = false
tenant_id = "{{foo.bar.baz}}"

[sinks.loki_receiver.encoding]
codec = "json"

[sinks.loki_receiver.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

# Basic Auth Config
[sinks.loki_receiver.auth]
strategy = "basic"
user = "username"
password = "password"

`,
		}),
		Entry("with custom bearer token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "http://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						Secret: &logging.OutputSecretSpec{
							Name: "custom-loki-secret",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"token": []byte("token-for-custom-loki"),
					},
				},
			},
			ExpectedConf: `
[transforms.loki_receiver_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_remap"]
endpoint = "http://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.loki_receiver.encoding]
codec = "json"

[sinks.loki_receiver.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

# Bearer Auth Config
[sinks.loki_receiver.auth]
strategy = "bearer"
token = "token-for-custom-loki"
`,
		}),
	)
})

var _ = Describe("Generate vector config for in cluster loki", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[constants.LogCollectorToken], generator.NoOptions)
	}
	DescribeTable("for Loki output", helpers.TestGenerateConfWith(f),
		Entry("with default logcollector bearer token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "http://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				constants.LogCollectorToken: {
					Data: map[string][]byte{
						"token": []byte("token-for-internal-loki"),
					},
				},
			},
			ExpectedConf: `
[transforms.loki_receiver_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_remap"]
endpoint = "http://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.loki_receiver.encoding]
codec = "json"

[sinks.loki_receiver.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

[sinks.loki_receiver.tls]
enabled = true
ca_file = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
# Bearer Auth Config
[sinks.loki_receiver.auth]
strategy = "bearer"
token = "token-for-internal-loki"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}

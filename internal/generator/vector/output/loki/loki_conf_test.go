package loki

import (
	"sort"
	"strings"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/tls"

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
	defaultTLS := "VersionTLS12"
	defaultCiphers := "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], op)
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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
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
		Entry("with TLS insecureSkipVerify=true", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						Secret: &logging.OutputSecretSpec{
							Name: "custom-loki-secret",
						},
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"ca-bundle.crt": []byte("junk"),
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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
endpoint = "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"
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
verify_certificate = false
verify_hostname = false
ca_file = "/var/run/ocp-collector/secrets/custom-loki-secret/ca-bundle.crt"
`,
		}),
		Entry("with TLS insecureSkipVerify=true, no certificate in secret", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
endpoint = "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"
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
verify_certificate = false
verify_hostname = false
`,
		}),
		Entry("with TLS config with default minTLSVersion & ciphers", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
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
			Options: generator.Options{
				generator.MinTLSVersion: string(tls.DefaultMinTLSVersion),
				generator.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
			},
			ExpectedConf: `
[transforms.loki_receiver_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
endpoint = "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"
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
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"

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

[transforms.loki_receiver_dedot]
type = "lua"
inputs = ["loki_receiver_remap"]
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

[sinks.loki_receiver]
type = "loki"
inputs = ["loki_receiver_dedot"]
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

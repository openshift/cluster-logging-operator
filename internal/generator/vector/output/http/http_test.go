package http

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate vector config", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], generator.NoOptions)
	}
	DescribeTable("for Http output", helpers.TestGenerateConfWith(f),
		Entry("", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com",
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
						OutputTypeSpec: logging.OutputTypeSpec{
							Http: &logging.Http{
								Headers: map[string]string{
									"h2": "v2",
									"h1": "v1",
								},
								Method: "POST",
							},
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: `
[transforms.http_receiver_normalize_http]
type = "remap"
inputs = ["application"]
source = '''
  del(.file)
'''

[transforms.http_receiver_dedot]
type = "lua"
inputs = ["http_receiver_normalize_http"]
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

[sinks.http_receiver]
type = "http"
inputs = ["http_receiver_dedot"]
uri = "https://my-logstore.com"
method = "post"

[sinks.http_receiver.encoding]
codec = "json"

[sinks.http_receiver.request]
timeout_secs = 10
headers = {"h1"="v1","h2"="v2"}

# Basic Auth Config
[sinks.http_receiver.auth]
strategy = "basic"
user = "username"
password = "password"
`,
		}),
		Entry("with custom bearer token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com",
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
						OutputTypeSpec: logging.OutputTypeSpec{
							Http: &logging.Http{
								Headers: map[string]string{
									"h2": "v2",
									"h1": "v1",
								},
								Method: "POST",
							},
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"token": []byte("token-for-custom-http"),
					},
				},
			},
			ExpectedConf: `
[transforms.http_receiver_normalize_http]
type = "remap"
inputs = ["application"]
source = '''
  del(.file)
'''

[transforms.http_receiver_dedot]
type = "lua"
inputs = ["http_receiver_normalize_http"]
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

[sinks.http_receiver]
type = "http"
inputs = ["http_receiver_dedot"]
uri = "https://my-logstore.com"
method = "post"

[sinks.http_receiver.encoding]
codec = "json"

[sinks.http_receiver.request]
timeout_secs = 10
headers = {"h1"="v1","h2"="v2"}

# Bearer Auth Config
[sinks.http_receiver.auth]
strategy = "bearer"
token = "token-for-custom-http"
`,
		}),
		Entry("with Http config", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com",
						OutputTypeSpec: v1.OutputTypeSpec{Http: &v1.Http{
							Timeout: "50",
							Headers: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}},
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"token": []byte("token-for-custom-http"),
					},
				},
			},
			ExpectedConf: `
[transforms.http_receiver_normalize_http]
type = "remap"
inputs = ["application"]
source = '''
  del(.file)
'''

[transforms.http_receiver_dedot]
type = "lua"
inputs = ["http_receiver_normalize_http"]
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

[sinks.http_receiver]
type = "http"
inputs = ["http_receiver_dedot"]
uri = "https://my-logstore.com"
method = "post"

[sinks.http_receiver.encoding]
codec = "json"

[sinks.http_receiver.request]
timeout_secs = 50
headers = {"k1"="v1","k2"="v2"}

# Bearer Auth Config
[sinks.http_receiver.auth]
strategy = "bearer"
token = "token-for-custom-http"
`,
		}),
		Entry("with Http config", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com",
						OutputTypeSpec: v1.OutputTypeSpec{Http: &v1.Http{
							Timeout: "50",
							Headers: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}},
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"token": []byte("token-for-custom-http"),
					},
				},
			},
			ExpectedConf: `
[transforms.http_receiver_normalize_http]
type = "remap"
inputs = ["application"]
source = '''
  del(.file)
'''

[transforms.http_receiver_dedot]
type = "lua"
inputs = ["http_receiver_normalize_http"]
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

[sinks.http_receiver]
type = "http"
inputs = ["http_receiver_dedot"]
uri = "https://my-logstore.com"
method = "post"

[sinks.http_receiver.encoding]
codec = "json"

[sinks.http_receiver.request]
timeout_secs = 50
headers = {"k1"="v1","k2"="v2"}

# Bearer Auth Config
[sinks.http_receiver.auth]
strategy = "bearer"
token = "token-for-custom-http"
`,
		}),
		Entry("with TLS config", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com",
						OutputTypeSpec: v1.OutputTypeSpec{Http: &v1.Http{
							Timeout: "50",
							Headers: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}},
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"token":         []byte("token-for-custom-http"),
						"tls.crt":       []byte("-- crt-- "),
						"tls.key":       []byte("-- key-- "),
						"ca-bundle.crt": []byte("-- ca-bundle -- "),
					},
				},
			},
			ExpectedConf: `
[transforms.http_receiver_normalize_http]
type = "remap"
inputs = ["application"]
source = '''
  del(.file)
'''

[transforms.http_receiver_dedot]
type = "lua"
inputs = ["http_receiver_normalize_http"]
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

[sinks.http_receiver]
type = "http"
inputs = ["http_receiver_dedot"]
uri = "https://my-logstore.com"
method = "post"

[sinks.http_receiver.encoding]
codec = "json"

[sinks.http_receiver.request]
timeout_secs = 50
headers = {"k1"="v1","k2"="v2"}

[sinks.http_receiver.tls]
verify_certificate = false
verify_hostname = false
key_file = "/var/run/ocp-collector/secrets/http-receiver/tls.key"
crt_file = "/var/run/ocp-collector/secrets/http-receiver/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/http-receiver/ca-bundle.crt"

# Bearer Auth Config
[sinks.http_receiver.auth]
strategy = "bearer"
token = "token-for-custom-http"
`,
		}),
	)
})

func TestHeaders(t *testing.T) {
	h := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	expected := `{"k1"="v1","k2"="v2"}`
	got := utils.ToHeaderStr(h, "%q=%q")
	if got != expected {
		t.Logf("got: %s, expected: %s", got, expected)
		t.Fail()
	}
}

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}

package syslog

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("vector syslog clf output", func() {
	const (
		xyzDefaults = `
[transforms.syslog_xyz_dedot]
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

[transforms.syslog_xyz_json]
type = "remap"
inputs = ["syslog_xyz_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.syslog_xyz]
type = "socket"
inputs = ["syslog_xyz_json"]
address = "logserver:514"
mode = "xyz"

[sinks.syslog_xyz.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "user"
severity = "informational"
`
		tcpDefaults = `
[transforms.syslog_tcp_dedot]
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

[transforms.syslog_tcp_json]
type = "remap"
inputs = ["syslog_tcp_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.syslog_tcp]
type = "socket"
inputs = ["syslog_tcp_json"]
address = "logserver:514"
mode = "tcp"

[sinks.syslog_tcp.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "user"
severity = "informational"
`
		tlsInsecure = `
[transforms.syslog_tcp_dedot]
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

[transforms.syslog_tcp_json]
type = "remap"
inputs = ["syslog_tcp_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.syslog_tcp]
type = "socket"
inputs = ["syslog_tcp_json"]
address = "logserver:514"
mode = "tcp"

[sinks.syslog_tcp.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "user"
severity = "informational"

[sinks.syslog_tcp.tls]
enabled = true
verify_certificate = false
verify_hostname = false
`
		udpEverySetting = `
[transforms.syslog_udp_dedot]
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

[transforms.syslog_udp_json]
type = "remap"
inputs = ["syslog_udp_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.syslog_udp]
type = "socket"
inputs = ["syslog_udp_json"]
address = "logserver:514"
mode = "udp"

[sinks.syslog_udp.encoding]
codec = "syslog"
rfc = "rfc3164"
facility = "kern"
severity = "critical"
app_name = "appName"
msg_id = "msgID"
proc_id = "procID"
tag = "tag"
add_log_source = true
`

		tlsWithLogRecordReferences = `
[transforms.syslog_tls_dedot]
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

[transforms.syslog_tls_json]
type = "remap"
inputs = ["syslog_tls_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.syslog_tls]
type = "socket"
inputs = ["syslog_tls_json"]
address = "logserver:6514"
mode = "tcp"

[sinks.syslog_tls.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "$$.message.facility"
severity = "$$.message.severity"
app_name = "$$.message.app_name"
msg_id = "$$.message.msg_id"
proc_id = "$$.message.proc_id"
tag = "$$.message.tag"
add_log_source = true

[sinks.syslog_tls.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/syslog-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/syslog-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/syslog-tls/ca-bundle.crt"
key_pass = "mysecretpassword"
`
	)

	var g generator.Generator

	var secrets = map[string]*corev1.Secret{
		"syslog-tls": {
			Data: map[string][]byte{
				"passphrase":    []byte("mysecretpassword"),
				"tls.key":       []byte("boo"),
				"tls.crt":       []byte("bar"),
				"ca-bundle.crt": []byte("baz"),
			},
		},
	}

	Context("syslog config", func() {
		BeforeEach(func() {
			g = generator.MakeGenerator()
		})

		It("LOG-4963: allow tls.insecureSkipVerify=true when no secret is defined", func() {
			element := Conf(
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-tcp",
					URL:  "tls://logserver:514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{},
					},
					TLS: &loggingv1.OutputTLSSpec{
						InsecureSkipVerify: true,
					},
				}, []string{"pipelineName"}, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(tlsInsecure))
		})
		It("LOG-3948: should pass URL scheme to vector for validation", func() {
			element := Conf(
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-xyz",
					URL:  "xyz://logserver:514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{},
					},
				}, []string{"pipelineName"}, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(xyzDefaults))
		})

		It("should configure TCP with defaults", func() {
			element := Conf(
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-tcp",
					URL:  "tcp://logserver:514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{},
					},
				}, []string{"pipelineName"}, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(tcpDefaults))
		})

		It("should configure UDP with every setting", func() {
			element := Conf(
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-udp",
					URL:  "udp://logserver:514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{
							RFC:          "rfc3164",
							Facility:     "kern",
							Severity:     "critical",
							AppName:      "appName",
							MsgID:        "msgID",
							ProcID:       "procID",
							Tag:          "tag",
							AddLogSource: true,
						},
					},
				}, []string{"pipelineName"}, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(udpEverySetting))
		})

		It("should configure TLS with log record field references", func() {
			element := Conf(
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-tls",
					URL:  "tls://logserver:6514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{
							RFC:          "rfc5424",
							Facility:     "$.message.facility",
							Severity:     "$.message.severity",
							AppName:      "$.message.app_name",
							MsgID:        "$.message.msg_id",
							ProcID:       "$.message.proc_id",
							Tag:          "$.message.tag",
							AddLogSource: true,
						},
					},
					Secret: &loggingv1.OutputSecretSpec{
						Name: "syslog-tls",
					},
				}, []string{"pipelineName"}, secrets["syslog-tls"], nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(tlsWithLogRecordReferences))
		})

	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vector syslog conf generation")
}

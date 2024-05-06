package syslog

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("vector syslog clf output", func() {
	const (
		xyzDefaults = `
[transforms.example_dedot]
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
[transforms.example_json]
type = "remap"
inputs = ["example_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.example]
type = "socket"
inputs = ["example_json"]
address = "logserver:514"
mode = "xyz"

[sinks.example.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "user"
severity = "informational"
`
		tcpDefaults = `
[transforms.example_dedot]
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

[transforms.example_json]
type = "remap"
inputs = ["example_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.example]
type = "socket"
inputs = ["example_json"]
address = "logserver:514"
mode = "tcp"

[sinks.example.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "user"
severity = "informational"
`
		tlsInsecure = `
[transforms.example_dedot]
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

[transforms.example_json]
type = "remap"
inputs = ["example_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.example]
type = "socket"
inputs = ["example_json"]
address = "logserver:514"
mode = "tcp"

[sinks.example.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "user"
severity = "informational"

[sinks.example.tls]

verify_certificate = false
verify_hostname = false
`
		udpEverySetting = `
[transforms.example_dedot]
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

[transforms.example_json]
type = "remap"
inputs = ["example_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.example]
type = "socket"
inputs = ["example_json"]
address = "logserver:514"
mode = "udp"

[sinks.example.encoding]
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
[transforms.example_dedot]
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

[transforms.example_json]
type = "remap"
inputs = ["example_dedot"]
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.example]
type = "socket"
inputs = ["example_json"]
address = "logserver:6514"
mode = "tcp"

[sinks.example.encoding]
codec = "syslog"
rfc = "rfc5424"
facility = "$$.message.facility"
severity = "$$.message.severity"
app_name = "$$.message.app_name"
msg_id = "$$.message.msg_id"
proc_id = "$$.message.proc_id"
tag = "$$.message.tag"
add_log_source = true

[sinks.example.tls]

key_file = "/var/run/ocp-collector/secrets/syslog-tls/tls.key"
crt_file = "/var/run/ocp-collector/secrets/syslog-tls/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/syslog-tls/ca-bundle.crt"
key_pass = "mysecretpassword"
`
	)

	var g framework.Generator

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
		const outputName = "example"
		BeforeEach(func() {
			g = framework.MakeGenerator()
		})

		It("LOG-4963: allow tls.insecureSkipVerify=true when no secret is defined", func() {
			element := New(
				outputName,
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
				}, []string{"pipelineName"}, nil, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(tlsInsecure))
		})

		It("LOG-3948: should pass URL scheme to vector for validation", func() {
			element := New(
				outputName,
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-xyz",
					URL:  "xyz://logserver:514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{},
					},
				}, []string{"pipelineName"}, nil, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(xyzDefaults))
		})

		It("should configure TCP with defaults", func() {
			element := New(
				outputName,
				loggingv1.OutputSpec{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-tcp",
					URL:  "tcp://logserver:514",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Syslog: &loggingv1.Syslog{},
					},
				}, []string{"pipelineName"}, nil, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(tcpDefaults))
		})

		It("should configure UDP with every setting", func() {
			element := New(
				outputName,
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
				}, []string{"pipelineName"}, nil, nil, nil)
			results, err := g.GenerateConf(element...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(udpEverySetting))
		})

		It("should configure TLS with log record field references", func() {
			element := New(
				outputName,
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
				}, []string{"pipelineName"}, secrets["syslog-tls"], nil, nil)
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

package syslog

import (
	_ "embed"
	corev1 "k8s.io/api/core/v1"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

//go:embed xyz_defaults.toml
var xyzDefaults string

//go:embed tls_with_field_references.toml
var tlsWithLogRecordReferences string

//go:embed udp_with_every_setting.toml
var udpEverySetting string

//go:embed tcp_with_defaults.toml
var tcpDefaults string

//go:embed tls_insecure.toml
var tlsInsecure string

var _ = Describe("vector syslog clf output", func() {

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
							AddLogSource: true,
						},
					},
					Secret: &loggingv1.OutputSecretSpec{
						Name: "syslog-tls",
					},
				}, []string{"pipelineName"}, secrets["syslog-tls"], nil, nil)
			results, err := g.GenerateConf(element...)
			println(results)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(tlsWithLogRecordReferences))
		})

	})
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vector syslog conf generation")
}

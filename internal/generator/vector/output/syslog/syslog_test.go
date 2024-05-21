package syslog_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/syslog"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("vector syslog clf output", func() {

	const (
		secretName = "syslog-tls"
	)

	var (
		tlsSpec = &obs.OutputTLSSpec{
			TLSSpec: obs.TLSSpec{
				CA: &obs.ConfigMapOrSecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.TrustedCABundleKey,
				},
				Certificate: &obs.ConfigMapOrSecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.ClientCertKey,
				},
				Key: &obs.SecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.ClientPrivateKey,
				},
				KeyPassphrase: &obs.SecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.Passphrase,
				},
			},
		}
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeSyslog,
				Name: "example",
				Syslog: &obs.Syslog{
					RFC: obs.SyslogRFC5424,
					URLSpec: obs.URLSpec{
						URL: "xyz://logserver:514",
					},
				},
			}
		}
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					"passphrase":    []byte("mysecretpassword"),
					"tls.key":       []byte("boo"),
					"tls.crt":       []byte("bar"),
					"ca-bundle.crt": []byte("baz"),
				},
			},
		}
	)
	DescribeTable("#New", func(expFile string, visit func(spec *obs.OutputSpec)) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		conf := syslog.New(outputSpec.Name, outputSpec, []string{"application"}, secrets, nil, framework.NoOptions)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("LOG-3948: should pass URL scheme to vector for validation", "xyz_defaults.toml", nil),
		Entry("should configure TCP with defaults", "tcp_with_defaults.toml", func(spec *obs.OutputSpec) {
			spec.Syslog.URL = "tcp://logserver:514"
		}),
		Entry("should configure UDP with every setting", "udp_with_every_setting.toml", func(spec *obs.OutputSpec) {
			spec.Syslog = &obs.Syslog{
				URLSpec:  obs.URLSpec{URL: "udp://logserver:514"},
				RFC:      obs.SyslogRFC3164,
				Facility: "kern",
				Severity: "critical",
				AppName:  "appName",
				MsgID:    "msgID",
				ProcID:   "procID",
			}
		}),
		Entry("should configure TLS with log record field references", "tls_with_field_references.toml", func(spec *obs.OutputSpec) {
			spec.TLS = tlsSpec
			spec.Syslog = &obs.Syslog{
				URLSpec:    obs.URLSpec{URL: "tls://logserver:6514"},
				RFC:        obs.SyslogRFC5424,
				Facility:   "$.message.facility",
				Severity:   "$.message.severity",
				AppName:    "$.message.app_name",
				MsgID:      "$.message.msg_id",
				ProcID:     "$.message.proc_id",
				PayloadKey: "$.message",
			}
		}),
	)

})

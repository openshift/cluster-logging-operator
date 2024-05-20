package splunk_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/splunk"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating vector config for Splunk output", func() {
	const (
		hecToken   = "VS0BNth3wCGF0eol0MuK07SHIrhYwCPHFWMG"
		secretName = "vector-splunk-secret"
		aToken     = "atoken"
	)
	var (
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					constants.TokenKey: []byte(aToken),
					"hecToken":         []byte(hecToken),
					"tls.key":          []byte("junk"),
					"tls.crt":          []byte("junk"),
					"ca-bundle.crt":    []byte("junk"),
					"passphrase":       []byte("junk"),
				},
			},
		}
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
			},
		}
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeSplunk,
				Name: "splunk_hec",
				Splunk: &obs.Splunk{
					URLSpec: obs.URLSpec{URL: "https://splunk-web:8088/endpoint"},
					Authentication: &obs.SplunkAuthentication{
						Token: &obs.SecretKey{
							Secret: &corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: constants.SplunkHECTokenKey,
						},
					},
					IndexSpec: obs.IndexSpec{Index: "{{.log_type}}"},
				},
			}
		}
	)

	DescribeTable("#New", func(expFile string, op framework.Options, visit func(spec *obs.OutputSpec)) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		conf := splunk.New(helpers.MakeID(outputSpec.Name), outputSpec, []string{"pipelineName"}, secrets, nil, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with basic sink", "splunk_sink.toml", framework.NoOptions, nil),
		Entry("with tls spec", "splunk_sink_with_tls.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.TLS = tlsSpec
		}),
		Entry("with tls spec", "splunk_sink_with_tls_and_static_index.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.TLS = tlsSpec
			spec.Splunk.Index = "foo"
		}),
	)
})

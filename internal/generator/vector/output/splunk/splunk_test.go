package splunk_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/splunk"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Generating vector config for Splunk output", func() {
	const (
		hecToken   = "VS0BNth3wCGF0eol0MuK07SHIrhYwCPHFWMG"
		secretName = "vector-splunk-secret"
		aToken     = "atoken"
	)
	var (
		adapter *observability.Output
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
				CA: &obs.ValueReference{
					Key:        constants.TrustedCABundleKey,
					SecretName: secretName,
				},
				Certificate: &obs.ValueReference{
					Key:        constants.ClientCertKey,
					SecretName: secretName,
				},
				Key: &obs.SecretReference{
					Key:        constants.ClientPrivateKey,
					SecretName: secretName,
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
						Token: &obs.SecretReference{
							Key:        constants.SplunkHECTokenKey,
							SecretName: secretName,
						},
					},
				},
			}
		}
		baseTune = &obs.BaseOutputTuningSpec{
			DeliveryMode:     obs.DeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			MaxRetryDuration: utils.GetPtr(time.Duration(35)),
			MinRetryDuration: utils.GetPtr(time.Duration(20)),
		}
	)

	DescribeTable("#New", func(expFile string, op utils.Options, visit func(spec *obs.OutputSpec)) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		adapter = observability.NewOutput(outputSpec)
		conf := splunk.New(helpers.MakeID(outputSpec.Name), adapter, []string{"pipelineName"}, secrets, op)
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
		Entry("with custom static & dynamic index", "splunk_sink_with_custom_index.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.Splunk.Index = `foo-{.kubernetes.namespace_name||"missing"}`
		}),
		Entry("with custom static & dynamic index", "splunk_sink_with_custom_index_dedot.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.Splunk.Index = `foo-{.kubernetes.namespace_labels."test/logging.io"||"missing"}`
		}),
		Entry("with tuning", "splunk_tune.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.Splunk.Tuning = &obs.SplunkTuningSpec{
				BaseOutputTuningSpec: *baseTune,
				Compression:          "gzip",
			}
		}),
		Entry("with indexed fields", "splunk_sink_with_indexed_fields.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.Splunk.IndexedFields = []obs.FieldPath{`.log_source`, `.kubernetes.namespace_labels."bar/baz0-9.test"`, `.annotations."authorization.k8s.io/decision"`}
		}),
		Entry("with indexed fields & source", "splunk_sink_with_indexed_fields_and_source.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.Splunk.Source = `{.foo||"missing"}`
			spec.Splunk.IndexedFields = []obs.FieldPath{`.log_source`, `.kubernetes.namespace_labels."bar/baz0-9.test"`, `.annotations."authorization.k8s.io/decision"`}
		}),
		Entry("with payloadKey", "splunk_sink_payloadkey.toml", framework.NoOptions, func(spec *obs.OutputSpec) {
			spec.Splunk.PayloadKey = ".openshift"
		}))
})

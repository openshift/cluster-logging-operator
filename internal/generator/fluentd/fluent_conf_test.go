package fluentd

import (
	_ "embed"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:embed fluent_conf_test/ns_pipelines.conf
var ExpectedNSPipelinesConf string

//go:embed fluent_conf_test/pod_labels.conf
var ExpectedPodLabelsConf string

//go:embed fluent_conf_test/pod_labels_ns.conf
var ExpectedPodLabelsNSConf string

//go:embed fluent_conf_test/exclude.conf
var ExpectedExcludeConf string

//go:embed fluent_conf_test/well_formed.conf
var ExpectedWellFormedConf string

//go:embed fluent_conf_test/valid_forwarders.conf
var ExpectedValidForwardersConf string

var _ = Describe("Generating fluentd config", func() {
	var (
		forwarder  *logging.ClusterLogForwarderSpec
		g          generator.Generator
		op         generator.Options = generator.Options{generator.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)}
		secret     corev1.Secret
		secretData = map[string][]byte{
			"tls.key":       []byte("test-key"),
			"tls.crt":       []byte("test-crt"),
			"ca-bundle.crt": []byte("test-bundle"),
		}
		secrets = map[string]*corev1.Secret{
			"infra-es": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-infra-secret",
				},
				Data: secretData,
			},
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-other-secret",
				},
				Data: secretData,
			},
			"audit-es": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-audit-secret",
				},
				Data: secretData,
			},
		}
	)

	BeforeEach(func() {
		g = generator.MakeGenerator()
		op = generator.Options{generator.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)}
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "infra-es",
					URL:    "https://es.svc.infra.cluster:9999",
					Secret: &logging.OutputSecretSpec{Name: "my-infra-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-2",
					URL:  "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-other-secret",
					},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "audit-es",
					URL:  "https://es.svc.audit.cluster:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-audit-secret",
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "infra-pipeline",
					InputRefs:  []string{logging.InputNameInfrastructure},
					OutputRefs: []string{"infra-es"},
				},
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				},
				{
					Name:       "audit-pipeline",
					InputRefs:  []string{logging.InputNameAudit},
					OutputRefs: []string{"audit-es"},
				},
			},
		}
		secret = corev1.Secret{
			Data: map[string][]byte{
				"tls.key":       []byte(""),
				"tls.crt":       []byte(""),
				"ca-bundle.crt": []byte(""),
			},
		}
		secrets = map[string]*corev1.Secret{
			"infra-es":  &secret,
			"apps-es-1": &secret,
			"apps-es-2": &secret,
			"audit-es":  &secret,
		}
	})

	It("should generate fluent config for sending specific namespace logs to separate pipelines", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-1",
					URL:  "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-2",
					URL:  "https://es2.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name: "myInput",
					Application: &logging.Application{
						Namespaces: []string{"project1-namespace", "project2-namespace"},
					},
				},
				{
					Name: "myInput2",
					Application: &logging.Application{
						Namespaces: []string{"dev-apple", "project2-namespace"},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "my-default-pipeline",
					InputRefs:  []string{"application"},
					OutputRefs: []string{"apps-es-2"},
				},
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{"myInput"},
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				},
				{
					Name:       "apps-pipeline2",
					InputRefs:  []string{"myInput2", "application"},
					OutputRefs: []string{"apps-es-2"},
				},
			},
		}
		secrets = map[string]*corev1.Secret{
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
		}
		c := generator.MergeSections(Conf(nil, secrets, forwarder, constants.OpenshiftNS, constants.SingletonName, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())

		Expect(results).To(EqualTrimLines(ExpectedNSPipelinesConf))
	})

	It("should generate fluent config for sending only logs from pods identified by labels to output", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-2",
					URL:    "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-other-secret"},
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name: "myInput-1",
					Application: &logging.Application{
						Selector: &logging.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "production",
								"app":         "nginx",
							},
						},
					},
				},
				{
					Name: "myInput-2",
					Application: &logging.Application{
						Selector: &logging.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "dev",
								"app":         "nginx",
							},
						},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-prod-pipeline",
					InputRefs:  []string{"myInput-1"},
					OutputRefs: []string{"apps-es-1"},
				},
				{
					Name:       "apps-dev-pipeline",
					InputRefs:  []string{"myInput-2"},
					OutputRefs: []string{"apps-es-2"},
				},
				{
					Name:       "apps-default",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"default"},
				},
			},
		}
		secrets = map[string]*corev1.Secret{
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-other-secret",
				},
				Data: secretData,
			},
		}
		c := generator.MergeSections(Conf(nil, secrets, forwarder, constants.OpenshiftNS, constants.SingletonName, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(ExpectedPodLabelsConf))
	})

	It("should generate fluent config for sending only logs from pods identified by labels and namespaces to output", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-2",
					URL:    "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-other-secret"},
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name: "myInput-1",
					Application: &logging.Application{
						Selector: &logging.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "production",
								"app":         "nginx",
							},
						},
						Namespaces: []string{"project1-namespace"},
					},
				},
				{
					Name: "myInput-2",
					Application: &logging.Application{
						Selector: &logging.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "dev",
								"app":         "nginx",
							},
						},
						Namespaces: []string{"project2-namespace"},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-prod-pipeline",
					InputRefs:  []string{"myInput-1"},
					OutputRefs: []string{"apps-es-1"},
				},
				{
					Name:       "apps-dev-pipeline",
					InputRefs:  []string{"myInput-2"},
					OutputRefs: []string{"apps-es-2"},
				},
				{
					Name:       "apps-default",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"default"},
				},
			},
		}
		secrets = map[string]*corev1.Secret{
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-other-secret",
				},
				Data: secretData,
			},
		}
		c := generator.MergeSections(Conf(nil, secrets, forwarder, constants.OpenshiftNS, constants.SingletonName, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(ExpectedPodLabelsNSConf))
	})

	It("should exclude source to pipeline labels when there are no pipelines for a given sourceType (e.g. only logs.app)", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type: "fluentdForward",
					Name: "secureforward-receiver",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"secureforward-receiver"},
				},
			},
		}
		c := generator.MergeSections(Conf(nil, nil, forwarder, constants.OpenshiftNS, constants.SingletonName, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(ExpectedExcludeConf))
	})

	It("should produce well formed fluent.conf", func() {
		c := generator.MergeSections(Conf(nil, secrets, forwarder, constants.OpenshiftNS, constants.SingletonName, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(ExpectedWellFormedConf))
	})

	It("should generate sources for reserved inputs used as names or types", func() {
		sources := generator.GatherSources(&logging.ClusterLogForwarderSpec{
			Inputs: []logging.InputSpec{{Name: "in", Application: &logging.Application{}}},
			Pipelines: []logging.PipelineSpec{
				{
					InputRefs:  []string{"in"},
					OutputRefs: []string{"default"},
				},
				{
					InputRefs:  []string{"audit"},
					OutputRefs: []string{"default"},
				},
			},
		}, generator.NoOptions)
		Expect(sources.List()).To(ContainElements("application", "audit"))
	})

	Describe("Json Parsing", func() {
		fw := &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-1",
					URL:  "https://es.svc.messaging.cluster.local:9654",
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name:        "myInput",
					Application: &logging.Application{},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{"myInput"},
					OutputRefs: []string{"apps-es-1"},
					Parse:      "json",
				},
			},
		}

		It("should generate config for Json parsing", func() {
			c := PipelineToOutputs(fw, nil)
			spec, err := g.GenerateConf(c...)
			Expect(err).To(BeNil())
			Expect(spec).To(EqualTrimLines(`
# Copying pipeline apps-pipeline to outputs
<label @APPS_PIPELINE>
  # Parse the logs into json
  <filter /^(?!(kubernetes\.|)var\.log\.pods\.openshift-.+_|(kubernetes\.|)var\.log\.pods\.default_|(kubernetes\.|)var\.log\.pods\.kube-.+_|journal\.|system\.var\.log|linux-audit\.log|k8s-audit\.log|openshift-audit\.log|ovn-audit\.log).+/>
    @type parser
    key_name message
    reserve_data true
    hash_value_field structured
    emit_invalid_record_to_error false
    remove_key_name_field true
    <parse>
      @type json
      json_parser oj
    </parse>
  </filter>

  <match **>
    @type relabel
    @label @APPS_ES_1
  </match>
</label>
`))
		})
	})

	var _ = DescribeTable("Verify generated fluentd.conf for valid forwarder specs",
		func(yamlSpec, wantFluentdConf string) {
			var spec logging.ClusterLogForwarderSpec
			Expect(yaml.Unmarshal([]byte(yamlSpec), &spec)).To(Succeed())
			g := generator.MakeGenerator()
			s := Conf(nil, security.NoSecrets, &spec, constants.OpenshiftNS, constants.SingletonName, op)
			gotFluentdConf, err := g.GenerateConf(generator.MergeSections(s)...)
			Expect(err).To(Succeed())
			Expect(gotFluentdConf).To(EqualTrimLines(wantFluentdConf))
		},
		Entry("namespaces", `
pipelines:
  - name: test-app
    inputrefs:
      - inputest
    outputrefs:
      - default
inputs:
  - name: inputest
    application:
      namespaces:
       - project1
       - project2
        `,
			ExpectedValidForwardersConf))
})

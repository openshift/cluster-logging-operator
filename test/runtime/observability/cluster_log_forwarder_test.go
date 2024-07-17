package observability_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ClustLogForwarderBuilder", func() {
	var (
		forwarder *obs.ClusterLogForwarder
	)

	BeforeEach(func() {
		forwarder = &obs.ClusterLogForwarder{}
	})

	Context("#FromInput", func() {

		It("should correctly build with visitors", func() {
			NewClusterLogForwarderBuilder(forwarder).
				FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
					spec.Name = "custom-app"
					spec.Application.Tuning = &obs.ContainerInputTuningSpec{
						RateLimitPerContainer: &obs.LimitSpec{
							MaxRecordsPerSecond: 10,
						},
					}
				}).ToElasticSearchOutput()

			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`inputs:
- application:
    tuning:
      rateLimitPerContainer:
        maxRecordsPerSecond: 10
  name: custom-app
  type: application
outputs:
- name: elasticsearch
  elasticsearch:
    index: '{.log_type}-write'
    url: http://0.0.0.0:9200
  type: elasticsearch
pipelines:
- inputRefs:
  - custom-app
  name: forward-pipeline
  outputRefs:
  - elasticsearch
serviceAccount:
  name: ""
`))
		})
		It("should correctly build to multiple outputs", func() {
			pipelineBuilder := NewClusterLogForwarderBuilder(forwarder).
				FromInput(obs.InputTypeApplication)
			pipelineBuilder.ToElasticSearchOutput()
			pipelineBuilder.ToSyslogOutput(obs.SyslogRFC5424)

			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`inputs:
- application: {}
  name: application
  type: application
outputs:
- name: elasticsearch
  elasticsearch:
    index: '{.log_type}-write'
    url: http://0.0.0.0:9200
  type: elasticsearch
- name: syslog
  syslog:
    rfc: RFC5424
    url: tcp://127.0.0.1:24224
  type: syslog
pipelines:
- inputRefs:
  - application
  name: forward-pipeline
  outputRefs:
  - elasticsearch
  - syslog
serviceAccount:
  name: ""
`))
		})
	})

	Context("#FromInputWithVisitor", func() {

		It("should build from multiple inputs", func() {
			appLabels1 := map[string]string{"name": "app1", "env": "env1"}
			appLabels2 := map[string]string{"name": "app1", "fallback": "env2"}
			builder := NewClusterLogForwarderBuilder(forwarder).
				FromInputName("application-logs1",
					func(spec *obs.InputSpec) {
						spec.Type = obs.InputTypeApplication
						spec.Application = &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{
									Namespace: "abc",
								},
							},
							Excludes: []obs.NamespaceContainerSpec{
								{
									Namespace: "xyz",
								},
							},
							Selector: &v1.LabelSelector{
								MatchLabels: appLabels1,
							},
						}
					},
				).Named("app-1").
				ToHttpOutput()
			builder.FromInputName("application-logs2",
				func(spec *obs.InputSpec) {
					spec.Type = obs.InputTypeApplication
					spec.Application = &obs.Application{
						Selector: &v1.LabelSelector{
							MatchLabels: appLabels2,
						},
					}
				},
			).Named("app-2").
				ToOutputWithVisitor(
					func(spec *obs.OutputSpec) {
						spec.Type = obs.OutputTypeSyslog
						spec.Syslog = &obs.Syslog{
							RFC: obs.SyslogRFC5424,
							URL: "tcp://0.0.0.0:24225",
						}
					}, "other")
			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`inputs:
    - application:
        includes:
        - namespace: abc
        excludes:
        - namespace: xyz
        selector:
          matchLabels:
            env: env1
            name: app1
      name: application-logs1
      type: application
    - application:
        selector:
          matchLabels:
            fallback: env2
            name: app1
      name: application-logs2
      type: application
outputs:
- name: http
  http:
    method: POST
    url: http://localhost:8090
  type: http
- name: other
  type: syslog
  syslog:
    rfc: RFC5424
    url: tcp://0.0.0.0:24225
pipelines:
- inputRefs:
  - application-logs1
  name: app-1
  outputRefs:
  - http
- inputRefs:
  - application-logs2
  name: app-2
  outputRefs:
  - other
serviceAccount:
  name: ""
`))
		})
	})

	Context("#WithFilter", func() {
		It("should correctly build to Elasticsearch with prune filter", func() {
			builder := NewClusterLogForwarderBuilder(forwarder)
			builder.FromInput(obs.InputTypeApplication).
				WithFilter("foo-prune",
					func(spec *obs.FilterSpec) {
						spec.Type = obs.FilterTypePrune
						spec.PruneFilterSpec = &obs.PruneFilterSpec{
							NotIn: []obs.FieldPath{".log_type"},
						}
					}).ToElasticSearchOutput()

			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`filters:
- name: foo-prune
  prune:
    notIn:
    - .log_type
  type: prune
inputs:
- application: {}
  name: application
  type: application
outputs:
- name: elasticsearch
  elasticsearch:
    index: '{.log_type}-write'
    url: http://0.0.0.0:9200
  type: elasticsearch
pipelines:
- filterRefs:
  - foo-prune
  inputRefs:
  - application
  name: forward-pipeline
  outputRefs:
  - elasticsearch
serviceAccount:
  name: ""
`))
		})
	})

})

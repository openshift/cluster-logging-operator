package functional

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ClustLogForwarderBuilder", func() {

	var (
		forwarder *logging.ClusterLogForwarder
	)

	BeforeEach(func() {
		forwarder = &logging.ClusterLogForwarder{}
	})

	Context("#FromInput", func() {

		It("should correctly build to multiple outputs", func() {
			pipelineBuilder := NewClusterLogForwarderBuilder(forwarder).
				FromInput(logging.InputNameApplication)
			pipelineBuilder.ToElasticSearchOutput()
			pipelineBuilder.ToSyslogOutput()

			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`outputs:
- name: elasticsearch
  type: elasticsearch
  url: http://0.0.0.0:9200
- name: syslog
  type: syslog
  url: tcp://0.0.0.0:24224
pipelines:
- inputRefs:
  - application
  name: forward-pipeline
  outputRefs:
  - elasticsearch
  - syslog`))
		})
	})

	Context("#FromInputWithVisitor", func() {

		It("should build from multiple inputs", func() {
			appLabels1 := map[string]string{"name": "app1", "env": "env1"}
			appLabels2 := map[string]string{"name": "app1", "fallback": "env2"}
			builder := NewClusterLogForwarderBuilder(forwarder).
				FromInputWithVisitor("application-logs1",
					func(spec *logging.InputSpec) {
						spec.Application = &logging.Application{
							Selector: &metav1.LabelSelector{
								MatchLabels: appLabels1,
							},
						}
					},
				).Named("app-1").
				ToFluentForwardOutput()
			builder.FromInputWithVisitor("application-logs2",
				func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Selector: &metav1.LabelSelector{
							MatchLabels: appLabels2,
						},
					}
				},
			).Named("app-2").
				ToOutputWithVisitor(
					func(spec *logging.OutputSpec) {
						spec.Type = logging.OutputTypeFluentdForward
						spec.URL = "tcp://0.0.0.0:24225"
					}, "other")

			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`inputs:
- name: application-logs1
  application:
    selector:
      matchLabels: 
        env: env1
        name: app1
- name: application-logs2
  application:
    selector:
      matchLabels: 
        fallback: env2
        name: app1
outputs:
- name: fluentdForward
  type: fluentdForward
  url: tcp://0.0.0.0:24224
- name: other
  type: fluentdForward
  url: tcp://0.0.0.0:24225
pipelines:
- inputRefs:
  - application-logs1
  name: app-1
  outputRefs:
  - fluentdForward
- inputRefs:
  - application-logs2
  name: app-2
  outputRefs:
  - other
`))
		})
	})

})

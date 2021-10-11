package functional

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("ClustLogForwarderBuilder", func() {

	Context("#FromInput", func() {

		It("should correctly build to multiple outputs", func() {
			forwarder := &logging.ClusterLogForwarder{}
			pipelineBuilder := NewClusterLogForwarderBuilder(forwarder).
				FromInput(logging.InputNameApplication)
			pipelineBuilder.ToElasticSearchOutput()
			pipelineBuilder.ToSyslogOutput()

			Expect(test.YAMLString(forwarder.Spec)).To(MatchYAML(`outputs:
- name: elasticsearch
  type: elasticsearch
  url: https://0.0.0.0:9200
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
})

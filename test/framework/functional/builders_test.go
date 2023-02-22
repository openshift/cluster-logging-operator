package functional

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
)

var _ = Describe("CollectorFunctionalFrameworkUsingCollectorBuilder", func() {

	var (
		secret = runtime.NewSecret("", "mysecret", map[string][]byte{
			constants.ClientUsername: []byte("admin"),
			constants.ClientPassword: []byte("elasticadmin"),
		},
		)
	)

	BeforeEach(func() {
	})

	Context("#FromInput", func() {

		It("should correctly build to Elasticsearch", func() {
			builder := NewCollectorFunctionalFrameworkUsingCollectorBuilder("vector", client.DryRunTestOption)
			builder.FromInput(logging.InputNameApplication).
				ToElasticSearchOutput(ElasticsearchVersion7, secret)

			Expect(test.YAMLString(builder.Framework.Forwarder.Spec)).To(MatchYAML(`inputs:
- name: application
outputs:
- name: elasticsearch
  secret:
    name: mysecret
  elasticsearch:
    version: 7
  type: elasticsearch
  url: http://0.0.0.0:9200
pipelines:
- inputRefs:
  - application
  name: forward-pipeline
  outputRefs:
  - elasticsearch
`))
		})
	})
})

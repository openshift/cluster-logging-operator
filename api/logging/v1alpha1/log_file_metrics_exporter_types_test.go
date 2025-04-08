package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("LogFileMetricExporterSpec", func() {
	It("spec serialization", func() {
		lfmeSpec := &v1alpha1.LogFileMetricExporter{}
		specString := `
spec:
nodeselector:
  nodeKey: "nodeVal"
tolerations:
- key: "test"
  operator: "Exists"
  effect: "NoSchedule"
- key: "test2"
  operator: "Exists"
  efftect: "Scheduled"
resources:
  limits:
    cpu: "500m"
    memory: "100Mi"
  requests:
    cpu: "100m"
    memory: "5Gi"
`
		test.MustUnmarshal(specString, lfmeSpec)
		Expect(test.YAMLString(lfmeSpec), specString)
	})

})

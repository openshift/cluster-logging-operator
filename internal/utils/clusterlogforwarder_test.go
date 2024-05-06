package utils

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("ClusterLogForwarderUtils", func() {
	It("should return output map of name and the associated spec", func() {
		clfSpec := obs.ClusterLogForwarderSpec{
			Outputs: []obs.OutputSpec{
				{
					Name: "test-cw",
					Type: obs.OutputTypeCloudwatch,
				},
				{
					Name: "test-splunk",
					Type: obs.OutputTypeSplunk,
				},
				{
					Name: "test-es",
					Type: obs.OutputTypeElasticsearch,
				},
			},
		}
		expectedOutMap := map[string]*obs.OutputSpec{
			"test-cw":     {Name: "test-cw", Type: obs.OutputTypeCloudwatch},
			"test-splunk": {Name: "test-splunk", Type: obs.OutputTypeSplunk},
			"test-es":     {Name: "test-es", Type: obs.OutputTypeElasticsearch},
		}

		Expect(reflect.DeepEqual(expectedOutMap, OutputMap(&clfSpec))).To(BeTrue())
	})
})

package loki

import (
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("outputLabelConf", func() {
	var (
		loki *logging.Loki
	)
	BeforeEach(func() {
		loki = &logging.Loki{}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(defaultLabelKeys, requiredLabelKeys...)
				sort.Strings(exp)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})
		Context("are spec'd", func() {
			It("should use the ones provided and add the required ones", func() {
				loki.LabelKeys = []string{"foo"}
				exp := append(loki.LabelKeys, requiredLabelKeys...)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})

	})
})

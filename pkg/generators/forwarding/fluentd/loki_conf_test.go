package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"sort"
)

var _ = Describe("outputLabelConf", func() {
	var (
		conf *outputLabelConf
	)
	BeforeEach(func() {
		conf = &outputLabelConf{
			Target: logging.OutputSpec{
				OutputTypeSpec: logging.OutputTypeSpec{
					Loki: &logging.Loki{},
				},
			},
		}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(defaultLabelKeys, requiredLokiLabelKeys...)
				sort.Strings(exp)
				Expect(conf.lokiLabelKeys()).To(BeEquivalentTo(exp))
			})
		})
		Context("are spec'd", func() {
			It("should use the ones provided and add the required ones", func() {
				conf.Target.Loki.LabelKeys = []string{"foo"}
				exp := append(conf.Target.Loki.LabelKeys, requiredLokiLabelKeys...)
				Expect(conf.lokiLabelKeys()).To(BeEquivalentTo(exp))
			})
		})

	})
})

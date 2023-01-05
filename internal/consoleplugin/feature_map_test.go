package consoleplugin

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#FeaturesForOCP", func() {
	It("should return the default set when there is no version match ", func() {
		Expect(FeaturesForOCP("4.12.0-rc.3")).To(Equal(featuresIfUnmatched))
	})

	It("should not enable the dev console for OCP 4.10", func() {
		Expect(FeaturesForOCP("4.10.4+abc")).To(Equal([]string{}))
	})
})

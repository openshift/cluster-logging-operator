package console

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#FeaturesForOCP", func() {
	It("should return an empty set when there is an empty version ", func() {
		Expect(FeaturesForOCP("")).To(Equal([]string{}))
	})
	It("should return the default set when there is no version match ", func() {
		Expect(FeaturesForOCP("4.12.0-rc.3")).To(Equal(featuresIfUnmatched))
	})

	It("should return the default set when greater than or equal 4.11", func() {
		Expect(FeaturesForOCP("4.11.12")).To(Equal(featuresIfUnmatched))
	})

	It("should not enable the dev console for OCP 4.10", func() {
		Expect(FeaturesForOCP("4.10.4+abc")).To(Equal([]string{}))
	})

	It("should not enable the dev console for OCP 4.10.51", func() {
		Expect(FeaturesForOCP("4.10.51")).To(Equal([]string{}))
	})
})

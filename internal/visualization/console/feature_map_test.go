package console

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#FeaturesForOCP", func() {
	It("should return an empty set when there is an empty version ", func() {
		Expect(FeaturesForOCP("")).To(Equal([]string{}))
	})
	It("should enable the the dev console and alerts when <= 4.13", func() {
		features := []string{featureAlerts, featureDevConsole}
		Expect(FeaturesForOCP("4.13.0-0.nightly-2023-03-23-204038")).To(Equal(features))
		Expect(FeaturesForOCP("4.13.0")).To(Equal(features))
		Expect(FeaturesForOCP("4.14.0")).To(Equal(features))
	})
	It("should enable the dev console when <= 4.11 but < 4.13", func() {
		features := []string{featureDevConsole}
		Expect(FeaturesForOCP("4.11.12")).To(Equal(features))
		Expect(FeaturesForOCP("4.12.0-rc.3")).To(Equal(features))
	})

	It("should not enable the dev console for OCP 4.10", func() {
		Expect(FeaturesForOCP("4.10.4+abc")).To(Equal([]string{}))
	})

	It("should not enable the dev console for OCP 4.10.51", func() {
		Expect(FeaturesForOCP("4.10.51")).To(Equal([]string{}))
	})
})

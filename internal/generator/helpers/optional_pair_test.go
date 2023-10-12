package helpers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OptionalPair", func() {
	Context("#Template", func() {
		var (
			pair OptionalPair
		)
		It("should return an empty template when the value is nil", func() {
			pair = NewOptionalPair("abc", nil)
			Expect(pair.String()).To(BeEmpty())
		})
		It("should return a formatted string config when value is a string", func() {
			pair = NewOptionalPair("abc", "xyz")
			Expect(pair.String()).To(Equal(`abc = "xyz"`))
		})
		It("should return a formatted numerical config when value is an int", func() {
			pair = NewOptionalPair("abc", 123)
			Expect(pair.String()).To(Equal(`abc = 123`))
		})
		It("should return a formatted bool config when value is bool", func() {
			pair = NewOptionalPair("abc", true)
			Expect(pair.String()).To(Equal(`abc = true`))
		})
		It("should return a formatted false bool config when value is bool", func() {
			pair = NewOptionalPair("abc", false)
			Expect(pair.String()).To(Equal(`abc = false`))
		})
	})
})

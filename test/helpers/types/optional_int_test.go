package types

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("OptionalInt", func() {
	Context("#NewOptionalInt", func() {
		It("should return an empty value for empty strings", func() {
			Expect(NewOptionalInt("  ")).To(Equal(OptionalInt("")))
		})
		It("should init the type with the value", func() {
			Expect(NewOptionalInt(">6")).To(Equal(OptionalInt(">6")))
		})
	})
	DescribeTable("#IsSatisfiedBy", func(expOptInt, otherValue string, satisfied bool) {
		exp := NewOptionalInt(expOptInt)
		other := NewOptionalInt(otherValue)
		Expect(exp.IsSatisfiedBy(other)).To(Equal(satisfied), "Exp. %q to satisfy %q", otherValue, expOptInt)
	},
		Entry("GreaterThan ", ">6", "7", true),
		Entry("GreaterThan Equal ", ">=6", "6", true),
		Entry("GreaterThan Equal ", ">=6", "7", true),
		Entry("Explicit Equal ", "=6", "6", true),
		Entry("Equal ", "6", "6", true),
		Entry("LessThan ", "<6", "5", true),
		Entry("LessThan Equal ", "<=6", "6", true),
		Entry("LessThan Equal ", "<=6", "5", true),
		Entry("Fail LessThan Equal ", "<=6", "8", false),
		Entry("Equal empty values", "", "", true),
		Entry("Optional provided but not required", "", "6", true),
	)
})

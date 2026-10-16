package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("#ValidateTOMLLiteralSafe", func() {
	It("should accept a benign value", func() {
		Expect(ValidateTOMLLiteralSafe("field", "app-logs-{.log_type}")).To(BeEmpty())
	})
	It("should accept a value with a single quote (not the terminator)", func() {
		Expect(ValidateTOMLLiteralSafe("field", "it's-fine")).To(BeEmpty())
	})
	It("should reject the literal multiline terminator", func() {
		Expect(ValidateTOMLLiteralSafe("field", "x'''injection")).To(ContainSubstring("must not contain the sequence"))
	})
	It("should reject newlines", func() {
		Expect(ValidateTOMLLiteralSafe("field", "line1\nline2")).To(ContainSubstring("newlines"))
	})
	It("should reject carriage returns", func() {
		Expect(ValidateTOMLLiteralSafe("field", "line1\rline2")).To(ContainSubstring("newlines"))
	})
})

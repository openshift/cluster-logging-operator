package test_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Matchers", func() {
	It("matches by line", func() {
		Expect(` foo
bar

baz
`).To(EqualTrimLines(`foo
	bar
baz`))
	})
	It("fails to match by line", func() {
		Expect("a").NotTo(EqualTrimLines("b"))
	})
})

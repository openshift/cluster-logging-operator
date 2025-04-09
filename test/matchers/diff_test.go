package matchers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
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

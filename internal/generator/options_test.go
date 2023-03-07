package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
)

var _ = Describe("Options#Has", func() {

	var (
		options = Options{}
	)
	It("should be false when the key does not exist", func() {
		Expect(options.Has("foo")).To(BeFalse())
	})

	It("should be true when the key exists", func() {
		options["mykey"] = 18
		Expect(options.Has("mykey")).To(BeTrue())
	})

})

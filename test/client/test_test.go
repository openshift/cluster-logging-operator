package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/openshift/cluster-logging-operator/test/client"
)

var _ = Describe("TestOptions", func() {

	var (
		options TestOptions
	)

	BeforeEach(func() {
		options = TestOptions{}
	})

	Context("#Includes", func() {
		It("should return false when the option is missing from the list", func() {
			Expect(options.Include(UseInfraNamespaceTestOption)).To(BeFalse())
		})

		It("should return true when the list contains the option", func() {
			options = append(options, UseInfraNamespaceTestOption)
			Expect(options.Include(UseInfraNamespaceTestOption)).To(BeTrue())
		})
	})
})

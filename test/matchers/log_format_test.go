package matchers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Log Format matcher tests", func() {

	It("match same value", func() {
		Expect(types.AllLog{Message: "text"}).To(FitLogFormatTemplate(types.AllLog{Message: "text"}))
	})

	It("match any (not Nil) value", func() {
		Expect(types.AllLog{Message: "text"}).To(FitLogFormatTemplate(types.AllLog{Message: "*"}))
	})

	It("match same nested value", func() {
		Expect(types.AllLog{Docker: types.Docker{ContainerID: "text"}}).To(FitLogFormatTemplate(types.AllLog{Docker: types.Docker{ContainerID: "text"}}))
	})

	It("do not match wrong value", func() {
		Expect(types.AllLog{Message: "wrong"}).NotTo(FitLogFormatTemplate(types.AllLog{Message: "text"}))
	})

	It("missing value do not match", func() {
		Expect(types.AllLog{}).NotTo(FitLogFormatTemplate(types.AllLog{Level: "text"}))
	})

	It("extra value do not match", func() {
		Expect(types.AllLog{Level: "text"}).NotTo(FitLogFormatTemplate(types.AllLog{}))
	})

	It("match regex", func() {
		Expect(types.AllLog{Message: "text"}).To(FitLogFormatTemplate(types.AllLog{Message: "regex:^[a-z]*$"}))
	})

	It("do not match wrong regex", func() {
		Expect(types.AllLog{Message: "text"}).NotTo(FitLogFormatTemplate(types.AllLog{Message: "regex:^[0-9]*$"}))
	})
})

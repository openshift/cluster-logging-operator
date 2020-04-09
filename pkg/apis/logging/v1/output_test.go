package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
)

var _ = Describe("Output", func() {
	It("recognizes output type names", func() {
		for _, name := range []string{"default", "syslog", "elasticsearch", "fluentForward"} {
			Expect(IsOutputTypeName(name)).To(BeTrue(), "IsOutputTypeName(%q)", name)
		}
	})
	It("rejects non-output type names", func() {
		for _, name := range []string{"foo", "", "xxx"} {
			Expect(IsOutputTypeName(name)).To(BeFalse(), "IsOutputTypeName(%q)", name)
		}
	})
	It("can lookup output type names by type", func() {
		Expect(OutputTypeName(outputs.Syslog{})).To(Equal("syslog"))
	})
	It("has values for output type names variables", func() {
		Expect(OutputTypeDefault).To(Equal("default"))
		Expect(OutputTypeFluent).To(Equal("fluentForward"))
	})
})

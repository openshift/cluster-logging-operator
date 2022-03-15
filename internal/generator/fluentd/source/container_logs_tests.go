package source

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("ContainerLogs", func() {
	var (
		conf = ContainerLogs{}
	)
	Context("#ReadLinesLimit", func() {
		It("should return nothing when there are no InFile tuning", func() {
			Expect(conf.ReadLinesLimit()).To(BeEmpty())
		})
		It("should return nothing when the InFile tuning is a negative number", func() {
			conf.Tunings = &logging.FluentdInFileSpec{ReadLinesLimit: -1}
			Expect(conf.ReadLinesLimit()).To(BeEmpty())
		})
		It("should return nothing when the InFile tuning is equal to zero", func() {
			conf.Tunings = &logging.FluentdInFileSpec{ReadLinesLimit: 0}
			Expect(conf.ReadLinesLimit()).To(BeEmpty())
		})
		It("should return a proper config parameter when the InFile tuning is greater then zero", func() {
			conf.Tunings = &logging.FluentdInFileSpec{ReadLinesLimit: 12}
			Expect(conf.ReadLinesLimit()).To(Equal("\n  read_lines_limit 12"))
		})
	})
})

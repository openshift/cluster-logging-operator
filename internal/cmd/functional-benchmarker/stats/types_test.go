package stats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaluating log loss stats", func() {

	var sample = `{"bloat":2.4790528233151186,"epoc_in":1699990366.4195192,"epoc_out":1699990381.2199013,"seqid":"0000000003"}`

	It("NewPerfLog", func() {
		log := NewPerfLog(sample, "loader-0.log")
		Expect(log.Bloat).To(BeNumerically("~", 2.4, 0.1))
		Expect(log.SequenceId).To(Equal(3))
		Expect(log.EpocIn).To(Equal(1699990366.4195192))
		Expect(log.EpocOut).To(Equal(1699990381.2199013))
		Expect(log.Stream).To(Equal("loader-0"))
	})
})

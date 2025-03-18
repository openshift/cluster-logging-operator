package stats

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaluating log loss stats", func() {

	var sample = `{"container_name":"loader-0","epoc_in":1742334136.8465972,"epoc_out":1742334154.0828154,"message":"goloader seq - functional.0000000000000000C08D75201BF50A40 - 0000000001","message_size":586,"payload_size":2616,"seqid":1}`

	It("NewPerfLog", func() {
		log := NewPerfLog(sample)
		Expect(log.Bloat()).To(BeNumerically("~", 4.4, 0.1))
		Expect(log.SequenceId).To(Equal(1))
		Expect(log.EpocIn).To(Equal(1742334136.8465972))
		Expect(log.EpocOut).To(Equal(1742334154.0828154))
		Expect(log.Stream).To(Equal("loader-0"))
	})
})

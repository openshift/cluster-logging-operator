package stats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaluating log loss stats", func() {

	var (
		logs = PerfLogs{
			{Stream: "host1", SequenceId: 2},
			{Stream: "host2", SequenceId: 1},
			{Stream: "host1", SequenceId: 3},
			{Stream: "host1", SequenceId: 4},
			{Stream: "host1", SequenceId: 10},
			{Stream: "host3", SequenceId: 1},
			{Stream: "host2", SequenceId: 2},
		}
		stats LossStats
	)

	BeforeEach(func() {
		stats = NewLossStats(logs)
	})

	Context("#splitEntriesByLoader", func() {
		It("should separate logs by stream", func() {
			Expect(splitEntriesByLoader(logs)).To(Equal(map[string][]PerfLog{
				"host1": {
					{Stream: "host1", SequenceId: 2},
					{Stream: "host1", SequenceId: 3},
					{Stream: "host1", SequenceId: 4},
					{Stream: "host1", SequenceId: 10},
				},
				"host2": {
					{Stream: "host2", SequenceId: 1},
					{Stream: "host2", SequenceId: 2},
				},
				"host3": {
					{Stream: "host3", SequenceId: 1},
				},
			}))
		})
	})

	Context("#Streams", func() {
		It("should return the list of hosts", func() {
			Expect(stats.Streams()).To(Equal([]string{"host1", "host2", "host3"}))
		})
	})
	Context("#LossStatsFor", func() {
		It("should return the total number of missing entries", func() {
			streamStats, err := stats.LossStatsFor("host1")
			Expect(err).To(BeNil())
			Expect(streamStats.Collected).To(Equal(4), "Exp to find calculate total missing from the stream")
			Expect(streamStats.Range()).To(Equal(8), "Exp to find the total possible entries for the stream")
		})
	})

})

package stats

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

var _ = Describe("Evaluating log loss stats", func() {

	var (
		formatLoaderMessage = func(seq int) string {
			return fmt.Sprintf("goloader seq - %s - %010d - %s", "functional.0.00000000000000009B38CE8D200310A4", seq, "some message")
		}

		newAllLog = func(containerName string, seq int) types.AllLog {
			return types.AllLog{
				ContainerLog: types.ContainerLog{
					Kubernetes: types.Kubernetes{ContainerName: containerName},
					ViaQCommon: types.ViaQCommon{Message: formatLoaderMessage(seq)},
				},
			}
		}

		logs = PerfLogs{
			{AllLog: newAllLog("host1", 2)},
			{AllLog: newAllLog("host2", 1)},
			{AllLog: newAllLog("host1", 3)},
			{AllLog: newAllLog("host1", 4)},
			{AllLog: newAllLog("host1", 10)},
			{AllLog: newAllLog("host3", 1)},
			{AllLog: newAllLog("host2", 2)},
		}

		baselineLogs = PerfLogs{
			{AllLog: newAllLog("", 2)},
			{AllLog: newAllLog("", 1)},
			{AllLog: newAllLog("", 3)},
			{AllLog: newAllLog("", 4)},
			{AllLog: newAllLog("", 10)},
			{AllLog: newAllLog("", 1)},
			{AllLog: newAllLog("", 2)},
		}

		stats LossStats
	)

	BeforeEach(func() {
		stats = NewLossStats(logs)
	})

	Context("#splitEntriesByLoader", func() {

		It("should separate logs by info in the message for baseline messages", func() {
			Expect(splitEntriesByLoader(baselineLogs)).To(Equal(map[string][]PerfLog{
				"functional.0.00000000000000009B38CE8D200310A4": {
					{AllLog: newAllLog("", 2)},
					{AllLog: newAllLog("", 1)},
					{AllLog: newAllLog("", 3)},
					{AllLog: newAllLog("", 4)},
					{AllLog: newAllLog("", 10)},
					{AllLog: newAllLog("", 1)},
					{AllLog: newAllLog("", 2)},
				},
			}))
		})
		It("should separate logs by host", func() {
			Expect(splitEntriesByLoader(logs)).To(Equal(map[string][]PerfLog{
				"host1": {
					{AllLog: newAllLog("host1", 2)},
					{AllLog: newAllLog("host1", 3)},
					{AllLog: newAllLog("host1", 4)},
					{AllLog: newAllLog("host1", 10)},
				},
				"host2": {
					{AllLog: newAllLog("host2", 1)},
					{AllLog: newAllLog("host2", 2)},
				},
				"host3": {
					{AllLog: newAllLog("host3", 1)},
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

	Context("#GetSequenceIdFrom", func() {
		It("should return the seq from the message", func() {
			Expect(GetSequenceIdFrom("goloader seq - functional.0.000000000000000020EDEA5A11C91A7C - 0000000006 - KXmfZDNuNWxJtHbhAciehWlkxYjRWrC")).To(Equal(6))
		})
		It("should return an error if format is unexpected", func() {
			_, err := GetSequenceIdFrom("here is my message")
			Expect(err).To(Not(BeNil()))
		})
	})

})

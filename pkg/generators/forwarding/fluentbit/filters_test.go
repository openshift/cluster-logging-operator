package fluentbit

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/luaunittest"
)

var testfunc = `
function reassemble_cri_logs(tag,timestamp,record)
end
`
var _ = Describe("fluentbit filters", func() {
	defer GinkgoRecover()
	Describe("#reassemble_cri_logs", func() {
		var (
			runner *luaunittest.Runner
			err    error
			input  luaunittest.Input
			record luaunittest.Record
			result luaunittest.Result
		)
		BeforeEach(func() {
			record = luaunittest.Record{
				"logtag":  "F",
				"message": "a message",
			}
			input = luaunittest.Input{
				Tag:       "var.log.container.foo",
				Timestamp: 88.0,
				Record:    record,
			}
			runner, err = luaunittest.NewRunner(
				ConcatCrioFilter,
				"reassemble_cri_logs",
				input,
			)
			Expect(err).To(BeNil(), "There should not be an error creating the test runner")
		})
		Context("when processing a full crio, unsplit log", func() {
			BeforeEach(func() {
				Expect(runner.Run()).To(BeNil())
				Expect(len(runner.Results) > 0).To(BeTrue(), "Exp. results from running the test")
				result = runner.Results[0]
			})

			It("should indicate the record is unmodifed", func() {
				Expect(result.Code).To(Equal(luaunittest.Unmodified))
			})
			It("should not modify the timestamp", func() {
				Expect(result.Timestamp).To(Equal(input.Timestamp))
			})
			It("should not modify the record", func() {
				Expect(result.Record).To(Equal(input.Record))
			})
		})
		Context("when processing a partial crio log", func() {
			BeforeEach(func() {
				record.SetParialRecord()
				Expect(runner.Run()).To(BeNil())
				Expect(len(runner.Results) > 0).To(BeTrue(), "Exp. results from running the test")
				result = runner.Results[0]
			})

			It("should indicate the record is dropped", func() {
				Expect(result.Code).To(Equal(luaunittest.Dropped))
			})
			It("should blank the timestamp", func() {
				Expect(result.Timestamp).To(Equal(0.0))
			})
			It("should blank the record", func() {
				Expect(result.Record).To(BeNil())
			})
		})
		Context("when processing records from a split crio log", func() {
			var (
				partialInput  luaunittest.Input
				separateInput luaunittest.Input
			)
			BeforeEach(func() {
				partialRecord := luaunittest.Record{
					"logtag":  "P",
					"message": "The begining ",
				}
				partialInput = luaunittest.Input{
					Tag:       "var.log.container.foo",
					Timestamp: 70.0,
					Record:    partialRecord,
				}
				separateInput = luaunittest.Input{
					Tag:       "var.log.container.foo",
					Timestamp: 70.0,
					Record: luaunittest.Record{
						"logtag":  "F",
						"message": "a new full record",
					},
				}
				runner.Inputs = append([]luaunittest.Input{partialInput}, runner.Inputs...)
				runner.Inputs = append(runner.Inputs, separateInput)
				Expect(runner.Run()).To(BeNil(), "Exp. no errors executing the test")
				Expect(len(runner.Results) == len(runner.Inputs)).To(BeTrue(), "Exp. results from running the test")
				result = runner.Results[0]
			})

			Context("the partial record", func() {
				It("should indicate the record is dropped", func() {
					Expect(result.Code).To(Equal(luaunittest.Dropped))
				})
			})
			Context("the full record", func() {
				BeforeEach(func() {
					result = runner.Results[1]
				})
				It("should indicate the record is modified", func() {
					Expect(result.Code).To(Equal(luaunittest.ModifiedRecord))
				})
				It("should keep the full log timestamp", func() {
					Expect(result.Timestamp).To(Equal(input.Timestamp))
				})
				It("should combine the message field of the records", func() {
					Expect(result.Record.Message()).To(
						Equal(strings.Join([]string{partialInput.Record.Message(), input.Record.Message()}, "")))
				})
			})
		})
	})
})

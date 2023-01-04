package matchers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"time"
)

var _ = Describe("Log Format matcher tests", func() {

	It("match same value", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "text"}}}).
			To(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "text"}}}))
	})

	It("match any (not Nil) value", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "text"}}}).
			To(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "*"}}}))
	})

	It("match same nested value", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{Docker: types.Docker{ContainerID: "text"}}}).
			To(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{Docker: types.Docker{ContainerID: "text"}}}))
	})

	It("do not match wrong value", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "wrong"}}}).
			To(Not(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "text"}}})))
	})

	It("missing value do not match", func() {
		Expect(types.AllLog{}).
			To(Not(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Level: "text"}}})))
	})

	It("extra value do not match", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Level: "text"}}}).
			To(Not(FitLogFormatTemplate(types.AllLog{})))
	})

	It("match regex", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "text"}}}).
			To(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "regex:^[a-z]*$"}}}))
	})

	It("do not match wrong regex", func() {
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "text"}}}).
			To(Not(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Message: "regex:^[0-9]*$"}}})))
	})

	It("match same time value", func() {
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Timestamp: nanoTime}}}).
			To(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Timestamp: nanoTime}}}))
	})

	It("match empty time value", func() {
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Timestamp: nanoTime}}}).
			To(FitLogFormatTemplate(types.AllLog{}))
	})
	Context("for optional ints", func() {
		It("it should pass when field is missing and value is optional", func() {
			Expect(types.AllLog{}).To(FitLogFormatTemplate(types.AllLog{
				ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt(""),
					},
				}}}))
		})
		It("it should pass when field exists and value is optional", func() {
			Expect(types.AllLog{
				ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt("5"),
					},
				}},
			}).To(FitLogFormatTemplate(types.AllLog{
				ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt(""),
					},
				}}}))
		})
		It("it should fail when field is missing and match expected", func() {
			Expect(types.AllLog{}).ToNot(FitLogFormatTemplate(types.AllLog{
				ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt("=8"),
					},
				}}}))
		})
		It("it should fail when field exists and value does not match spec", func() {
			Expect(types.AllLog{
				ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt("5"),
					},
				}}}).ToNot(FitLogFormatTemplate(types.AllLog{
				ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt("=8"),
					},
				}}}))
		})
	})

	It("do not match wrong time value", func() {
		timestamp1 := "2013-03-28T14:36:03.243000+00:00"
		nanoTime1, _ := time.Parse(time.RFC3339Nano, timestamp1)
		timestamp2 := "2014-04-28T14:36:03.243000+00:00"
		nanoTime2, _ := time.Parse(time.RFC3339Nano, timestamp2)
		Expect(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Timestamp: nanoTime1}}}).
			To(Not(FitLogFormatTemplate(types.AllLog{ContainerLog: types.ContainerLog{ViaQCommon: types.ViaQCommon{Timestamp: nanoTime2}}})))
	})

})

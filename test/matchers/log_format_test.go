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

	It("match same time value", func() {
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
		Expect(types.AllLog{Timestamp: nanoTime}).To(FitLogFormatTemplate(types.AllLog{Timestamp: nanoTime}))
	})

	It("match empty time value", func() {
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
		Expect(types.AllLog{Timestamp: nanoTime}).To(FitLogFormatTemplate(types.AllLog{}))
	})

	It("do not match wrong time value", func() {
		timestamp1 := "2013-03-28T14:36:03.243000+00:00"
		nanoTime1, _ := time.Parse(time.RFC3339Nano, timestamp1)
		timestamp2 := "2014-04-28T14:36:03.243000+00:00"
		nanoTime2, _ := time.Parse(time.RFC3339Nano, timestamp2)
		Expect(types.AllLog{Timestamp: nanoTime1}).NotTo(FitLogFormatTemplate(types.AllLog{Timestamp: nanoTime2}))
	})

})

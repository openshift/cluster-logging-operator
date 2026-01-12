package syslog

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[test][helpers][syslog] ParseRFC5424SyslogLogs", func() {

	const (
		testTime        = "2025-11-24T10:30:00.123456Z"
		expectedAppName = "eventrouter"
		expectedHost    = "my-test-node"
	)

	DescribeTable("should correctly parse valid RFC 5424 log lines",
		func(rawLogLine, expectedJSONPayload, expectedStructuredData string, expectedPriority int) {
			msg, err := ParseRFC5424SyslogLogs(rawLogLine)

			Expect(err).NotTo(HaveOccurred(), "Parsing should succeed without error")
			Expect(msg.Priority).To(Equal(expectedPriority), "Priority should match")
			Expect(msg.Version).To(Equal(1), "Version should be 1")
			Expect(msg.Hostname).To(Equal(expectedHost), "Hostname should match")
			Expect(msg.AppName).To(Equal(expectedAppName), "App-Name should match")
			Expect(msg.StructuredData).To(Equal(expectedStructuredData), "Structured Data should match")

			expectedTime, _ := time.Parse(time.RFC3339Nano, testTime)
			Expect(msg.Timestamp).To(BeTemporally("~", expectedTime, time.Second), "Timestamp should be correctly parsed")

		},
		Entry("when SD is hyphen and payload is simple JSON",
			fmt.Sprintf("<14>1 %s %s %s - - - %s", testTime, expectedHost, expectedAppName, `{"key":"value","event":"added"}`),
			`{"key":"value","event":"added"}`,
			"-",
			14,
		),
	)

	It("should return an error for malformed Syslog lines", func() {
		// A log line missing the version number
		malformedLog := fmt.Sprintf("<14>%s %s %s - - - %s", testTime, expectedHost, expectedAppName, `{"key":"value"}`)
		msg, err := ParseRFC5424SyslogLogs(malformedLog)
		Expect(err).To(HaveOccurred(), "Parsing should fail for a malformed line")
		Expect(msg).To(BeNil(), "Should not return any parsed messages")
		Expect(err.Error()).To(ContainSubstring("expected 9 submatches"), "Error message should indicate regex failure")
	})

	It("should return an error for badly formatted timestamps", func() {
		badTimeLog := fmt.Sprintf("<14>1 %s %s %s - - - %s", "24/11/2025 10:30", expectedHost, expectedAppName, `{"key":"value"}`)
		msg, err := ParseRFC5424SyslogLogs(badTimeLog)
		Expect(err).To(HaveOccurred(), "Parsing should fail due to bad timestamp format")
		Expect(msg).To(BeNil(), "Should not return any parsed messages")
	})
})

func TestSyslogParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Syslog Parser Unit Test Suite")
}

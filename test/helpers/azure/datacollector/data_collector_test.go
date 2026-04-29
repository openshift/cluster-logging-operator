package datacollector

import (
	_ "embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/mockoon"
)

//go:embed test_mockoon.log
var logData string

var _ = Describe("Parsing Mockoon logs for Azure Monitor (Data Collector API)", func() {

	It("should parse application logs from test data", func() {
		logs, err := extractStructuredLogs(logData, "application")
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(HaveLen(8))
	})

	It("should ignore non-POST requests", func() {
		input := mockoon.NewLogLine("GET", "/api/logs", 200, `[{"log_type":"application","message":"should be ignored"}]`)
		logs, err := extractStructuredLogs(input, "application")
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should ignore non-200 responses", func() {
		input := mockoon.NewLogLine("POST", "/api/logs", 500, `[{"log_type":"application","message":"should be ignored"}]`)
		logs, err := extractStructuredLogs(input, "application")
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should return empty for server-started-only output", func() {
		input := `{"app":"mockoon-server","level":"info","message":"Server started on port 3000","timestamp":"2024-02-21T10:45:23.445Z"}`
		logs, err := extractStructuredLogs(input, "application")
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should skip malformed body without error", func() {
		input := mockoon.NewLogLine("POST", "/api/logs", 200, `not-valid-json`)
		logs, err := extractStructuredLogs(input, "application")
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should aggregate logs across multiple batches", func() {
		line1 := mockoon.NewLogLine("POST", "/api/logs", 200,
			`[{"log_type":"application","message":"batch1-msg1"}]`)
		line2 := mockoon.NewLogLine("POST", "/api/logs", 200,
			`[{"log_type":"application","message":"batch2-msg1"},{"log_type":"infrastructure","message":"infra-msg"}]`)
		input := strings.Join([]string{line1, line2}, "\n")

		logs, err := extractStructuredLogs(input, "application")
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(HaveLen(2))
		Expect(logs[0].Message).To(Equal("batch1-msg1"))
		Expect(logs[1].Message).To(Equal("batch2-msg1"))
	})
})

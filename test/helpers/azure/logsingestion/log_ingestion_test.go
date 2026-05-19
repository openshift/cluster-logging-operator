package logsingestion

import (
	_ "embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/mockoon"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

//go:embed test_mockoon_log_ingestion.log
var logData string

var _ = Describe("Parsing Mockoon logs for Azure Log Ingestion API", func() {

	It("should parse application logs from test data", func() {
		logs, err := extractLogsAs[types.ApplicationLog](logData)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(HaveLen(3))

		expectedMessages := []string{"First test message", "Second test message", "Third test message"}
		for i, log := range logs {
			Expect(log.Message).To(Equal(expectedMessages[i]))
		}
	})

	It("should ignore OAuth2 token requests", func() {
		input := mockoon.NewLogLine("POST", "/test-tenant/oauth2/v2.0/token", 200,
			`[{"log_type":"application","message":"should be ignored"}]`)
		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should ignore GET requests", func() {
		input := mockoon.NewLogLine("GET", "/dataCollectionRules/dcr-test/streams/Custom-Test_CL", 204,
			`[{"log_type":"application","message":"should be ignored"}]`)
		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should ignore non-204 responses", func() {
		input := mockoon.NewLogLine("POST", "/dataCollectionRules/dcr-test/streams/Custom-Test_CL", 500,
			`[{"log_type":"application","message":"should be ignored"}]`)
		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should ignore requests to non-dataCollectionRules paths", func() {
		input := mockoon.NewLogLine("POST", "/some-other-endpoint", 204,
			`[{"log_type":"application","message":"should be ignored"}]`)
		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should return empty for server-started-only output", func() {
		input := `{"app":"mockoon-server","level":"info","message":"Server started on port 3000","timestamp":"2024-02-21T10:45:23.445Z"}`
		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should skip malformed body without error", func() {
		input := mockoon.NewLogLine("POST", "/dataCollectionRules/dcr-test/streams/Custom-Test_CL", 204, `not-valid-json`)
		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(BeEmpty())
	})

	It("should aggregate logs across multiple batches", func() {
		line1 := mockoon.NewLogLine("POST", "/dataCollectionRules/dcr-1/streams/Custom-A_CL", 204,
			`[{"log_type":"application","message":"batch1-msg1"},{"log_type":"application","message":"batch1-msg2"}]`)
		line2 := mockoon.NewLogLine("POST", "/dataCollectionRules/dcr-1/streams/Custom-A_CL", 204,
			`[{"log_type":"application","message":"batch2-msg1"}]`)
		input := strings.Join([]string{line1, line2}, "\n")

		logs, err := extractLogsAs[types.ApplicationLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(HaveLen(3))
		Expect(logs[0].Message).To(Equal("batch1-msg1"))
		Expect(logs[1].Message).To(Equal("batch1-msg2"))
		Expect(logs[2].Message).To(Equal("batch2-msg1"))
	})

	It("should parse audit logs", func() {
		input := mockoon.NewLogLine("POST", "/dataCollectionRules/dcr-audit/streams/Custom-Audit_CL", 204,
			`[{"log_type":"audit","verb":"get","requestURI":"/api/v1/namespaces","message":"audit msg 1"},`+
				`{"log_type":"audit","verb":"list","requestURI":"/api/v1/pods","message":"audit msg 2"}]`)
		logs, err := extractLogsAs[types.AuditLogCommon](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(HaveLen(2))
		Expect(logs[0].Verb).To(Equal("get"))
		Expect(logs[0].RequestURI).To(Equal("/api/v1/namespaces"))
		Expect(logs[1].Verb).To(Equal("list"))
		Expect(logs[1].RequestURI).To(Equal("/api/v1/pods"))
	})

	It("should parse infrastructure logs", func() {
		input := mockoon.NewLogLine("POST", "/dataCollectionRules/dcr-infra/streams/Custom-Infra_CL", 204,
			`[{"log_type":"infrastructure","message":"infra msg 1","level":"info","hostname":"node-1"},`+
				`{"log_type":"infrastructure","message":"infra msg 2","level":"warn","hostname":"node-2"}]`)
		logs, err := extractLogsAs[types.InfraLog](input)
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(HaveLen(2))
		Expect(logs[0].Message).To(Equal("infra msg 1"))
		Expect(logs[0].Level).To(Equal("info"))
		Expect(logs[0].Hostname).To(Equal("node-1"))
		Expect(logs[1].Message).To(Equal("infra msg 2"))
		Expect(logs[1].Level).To(Equal("warn"))
		Expect(logs[1].Hostname).To(Equal("node-2"))
	})
})

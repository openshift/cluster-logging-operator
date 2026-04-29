package gcl

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/mockoon"
)

//go:embed test_mockoon_gcl.log
var logData string

var _ = Describe("Parsing Mockoon logs for GCP Cloud Logging API", func() {

	It("should parse application logs from test data", func() {
		entries, err := extractRawEntries(logData)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(HaveLen(3))

		expectedMessages := []string{"First test message", "Second test message", "Third test message"}
		for i, raw := range entries {
			var payload map[string]interface{}
			Expect(json.Unmarshal([]byte(raw), &payload)).To(Succeed())
			Expect(payload["message"]).To(Equal(expectedMessages[i]))
		}
	})

	It("should ignore OAuth2 token requests", func() {
		body := `{"entries":[{"logName":"projects/test/logs/test","jsonPayload":{"log_type":"application","message":"should be ignored"},"severity":"DEFAULT"}]}`
		input := mockoon.NewLogLine("POST", "/token", 200, body)
		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should ignore GET requests", func() {
		body := `{"entries":[{"logName":"projects/test/logs/test","jsonPayload":{"log_type":"application","message":"should be ignored"},"severity":"DEFAULT"}]}`
		input := mockoon.NewLogLine("GET", "/v2/entries:write", 200, body)
		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should ignore non-200 responses", func() {
		body := `{"entries":[{"logName":"projects/test/logs/test","jsonPayload":{"log_type":"application","message":"should be ignored"},"severity":"DEFAULT"}]}`
		input := mockoon.NewLogLine("POST", "/v2/entries:write", 500, body)
		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should ignore requests to non-entries:write paths", func() {
		body := `{"entries":[{"logName":"projects/test/logs/test","jsonPayload":{"log_type":"application","message":"should be ignored"},"severity":"DEFAULT"}]}`
		input := mockoon.NewLogLine("POST", "/other-endpoint", 200, body)
		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should return empty for server-started-only output", func() {
		input := `{"app":"mockoon-server","level":"info","message":"Server started on port 3000","timestamp":"2024-02-21T10:45:23.445Z"}`
		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should skip malformed body without error", func() {
		input := mockoon.NewLogLine("POST", "/v2/entries:write", 200, `not-valid-json`)
		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should aggregate logs across multiple batches", func() {
		entry1 := gclEntryJSON("batch1-msg1")
		entry2 := gclEntryJSON("batch1-msg2")
		entry3 := gclEntryJSON("batch2-msg1")

		body1 := fmt.Sprintf(`{"entries":[%s,%s]}`, entry1, entry2)
		body2 := fmt.Sprintf(`{"entries":[%s]}`, entry3)

		line1 := mockoon.NewLogLine("POST", "/v2/entries:write", 200, body1)
		line2 := mockoon.NewLogLine("POST", "/v2/entries:write", 200, body2)
		input := strings.Join([]string{line1, line2}, "\n")

		entries, err := extractRawEntries(input)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(HaveLen(3))

		expectedMessages := []string{"batch1-msg1", "batch1-msg2", "batch2-msg1"}
		for i, raw := range entries {
			var payload map[string]interface{}
			Expect(json.Unmarshal([]byte(raw), &payload)).To(Succeed())
			Expect(payload["message"]).To(Equal(expectedMessages[i]))
		}
	})
})

func gclEntryJSON(message string) string {
	entry := GCLEntry{
		LogName: "projects/test-project/logs/test-log-id",
		JsonPayload: map[string]interface{}{
			"message":  message,
			"log_type": "application",
			"level":    "default",
		},
		Severity: "DEFAULT",
	}
	b, _ := json.Marshal(entry)
	return string(b)
}

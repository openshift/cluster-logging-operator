package http

import (
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[Functional][Outputs][HTTP] Logforwarding to VictoriaLogs", func() {

	var (
		framework *functional.CollectorFunctionalFramework

		// Template expected as output Log
		outputLogTemplate = functional.NewApplicationLogTemplate()
	)

	Context("should write to victorialogs", func() {
		DescribeTable("with custom headers", func(headers map[string]string) {
			outputLogTemplate.ViaqIndexName = "app-write"
			framework = functional.NewCollectorFunctionalFramework()
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToHttpOutput(func(output *obs.OutputSpec) {
					output.HTTP.URL = "http://0.0.0.0:9428/insert/jsonline"
					output.HTTP.LinePerEvent = true
					output.HTTP.Headers = headers
                        	})
			defer framework.Cleanup()
			Expect(framework.Deploy()).To(BeNil())
			timestamp := functional.CRIOTime(time.Now())
			ukr := "привіт "
			jp := "こんにちは "
			ch := "你好"
			msg := functional.NewCRIOLogMessage(timestamp, ukr+jp+ch, false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 10)).To(BeNil())
			Expect(framework.WriteMessagesWithNotUTF8SymbolsToLog()).To(BeNil())
			requestHeaders := map[string]string{
				"AccountID": "0",
				"ProjectID": "0",
			}
			for headerName := range requestHeaders {
				if v, ok := headers[headerName]; ok {
					requestHeaders[headerName] = v
				}
			}
			raw, err := framework.GetLogsFromVL(string(obs.OutputTypeHTTP), requestHeaders)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))
			// Parse log line
			var logs []map[string]string
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(len(logs)).To(Equal(11))
			//sort log by time before matching
			sort.Slice(logs, func(i, j int) bool {
				return strings.Compare(logs[i]["_time"], logs[j]["_time"]) < 0
			})

			Expect(logs[0]["_msg"]).To(Equal(ukr + jp + ch))
			Expect(logs[10]["_msg"]).To(Equal("������������"))
		},
			Entry("Non-default tenant ID", map[string]string{
				"AccountID":    "10",
				"ProjectID":    "10",
				"VL-Msg-Field": "message",
			}),
		)
	})
})

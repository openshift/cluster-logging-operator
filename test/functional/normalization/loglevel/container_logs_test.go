package loglevel

import (
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[functional][normalization][loglevel] tests for message format of container logs", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput().
			FromInput(obs.InputTypeAudit).
			ToHttpOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})
	DescribeTable("when evaluating an application message", func(expLevel, message string) {
		// Log message data
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

		// Template expected as output Log
		var outputLogTemplate = functional.NewApplicationLogTemplate()
		outputLogTemplate.TimestampLegacy = nanoTime
		outputLogTemplate.Timestamp = nanoTime
		outputLogTemplate.Message = message
		outputLogTemplate.Level = expLevel

		// Write log line
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	},
		Entry("should recognize an INFO message", "info", "I0920 14:22:00.089385       1 scheduler.go:592] \"Successfully bound pod to node\" pod=\"openshift-marketplace/community-operators-qrs99\" node=\"ip-10-0-215-216.us-east-2.compute.internal\" evaluatedNodes=6 feasibleNodes=3"),
		Entry("should recognize an INFO message(LOG-4480)", "info", "2023-08-28T19:53:52.115Z\tinfo\tconfiguration-policy-controller\tlogr@v1.2.2/logr.go:252\tProcessing the policies"),
		Entry("should recognize an ERROR message", "error", "E0427 02:47:01.619035 1 authentication.go:53] Unable to authenticate the request due to an error: invalid bearer token, context canceled"),
		Entry("should recognize an DEBUG message", "debug", "level=debug found the light"),
		Entry("should recognize an TRACE message(level)", "trace", "level=trace found the light"),
		Entry("should recognize an TRACE message(T[0-9])", "trace", "T1234 found the light"),
		Entry("should recognize a complex WARN message", "warn", `2022-01-26 13:41:57.149 mId:597c790b-6482-4a23-905c-49af6959241e cId:WEB::NAID-iOS-W42760F0-DE25-4622-90A0-4961356829A9-1643226116.796449-1643226117121::1643226117142::abunchofstuff::null::CnC::null::39.0.3::null::null::NONE::NONE::null::NONE WARN com.fake.services.controllers.ecom.BalanceController:84 - GETBULKBALANCE::RECV_REQ - 1|GetBulkBalanceRequest(cash=[GetBalanceRequest(barcode=--------, transactionNumber=---, storeNumber=873, registerID=0, startTime=2022-01-26T13:41:57, pin=null)]) | Headers-> ConsumerApp: STF & MessageId: 597c790b-6482-4a23-905c-49af6959241e & EndpointUrl: /v1/cash/getBalanceBulk & CorrelationId: null`),
		Entry("should recognize a complex INFO message", "info", `2022-01-26 13:41:57.149 mId:597c790b-6482-4a23-905c-49af6959241e cId:WEB::NAID-iOS-E42760F0-DE25-4622-90A0-4961356829A9-1643226116.796449-1643226117121::1643226117142::abunchofstuff::null::CnC::null::39.0.3::null::null::NONE::NONE::null::NONE INFO com.fake.services.controllers.ecom.BalanceController:84 - GETBULKBALANCE::RECV_REQ - 1|GetBulkBalanceRequest(cash=[GetBalanceRequest(barcode=--------, transactionNumber=---, storeNumber=873, registerID=0, startTime=2022-01-26T13:41:57, pin=null)]) | Headers-> ConsumerApp: STF & MessageId: 597c790b-6482-4a23-905c-49af6959241e & EndpointUrl: /v1/cash/getBalanceBulk & CorrelationId: null`),
		Entry("should recognize a complex ERROR message", "error", `2022-01-26 13:41:57.149 mId:597c790b-6482-4a23-905c-49af6959241e cId:WEB::NAID-iOS-I42760F0-DE25-4622-90A0-4961356829A9-1643226116.796449-1643226117121::1643226117142::abunchofstuff::null::CnC::null::39.0.3::null::null::NONE::NONE::null::NONE ERROR com.fake.services.controllers.ecom.BalanceController:84 - GETBULKBALANCE::RECV_REQ - 1|GetBulkBalanceRequest(cash=[GetBalanceRequest(barcode=--------, transactionNumber=---, storeNumber=873, registerID=0, startTime=2022-01-26T13:41:57, pin=null)]) | Headers-> ConsumerApp: STF & MessageId: 597c790b-6482-4a23-905c-49af6959241e & EndpointUrl: /v1/cash/getBalanceBulk & CorrelationId: null`),
		Entry("should recognize a complex DEBUG message", "debug", `2022-01-26 13:41:57.149 mId:597c790b-6482-4a23-905c-49af6959241e cId:WEB::NAID-iOS-D42760F0-DE25-4622-90A0-4961356829A9-1643226116.796449-1643226117121::1643226117142::abunchofstuff::null::CnC::null::39.0.3::null::null::NONE::NONE::null::NONE DEBUG com.fake.services.controllers.ecom.BalanceController:84 - GETBULKBALANCE::RECV_REQ - 1|GetBulkBalanceRequest(cash=[GetBalanceRequest(barcode=--------, transactionNumber=---, storeNumber=873, registerID=0, startTime=2022-01-26T13:41:57, pin=null)]) | Headers-> ConsumerApp: STF & MessageId: 597c790b-6482-4a23-905c-49af6959241e & EndpointUrl: /v1/cash/getBalanceBulk & CorrelationId: null`),
		Entry("should recognize a complex TRACE message", "trace", `2022-01-26 13:41:57.149 mId:597c790b-6482-4a23-905c-49af6959241e cId:WEB::NAID-iOS-D42760F0-DE25-4622-90A0-4961356829A9-1643226116.796449-1643226117121::1643226117142::abunchofstuff::null::CnC::null::39.0.3::null::null::NONE::NONE::null::NONE TRACE com.fake.services.controllers.ecom.BalanceController:84 - GETBULKBALANCE::RECV_REQ - 1|GetBulkBalanceRequest(cash=[GetBalanceRequest(barcode=--------, transactionNumber=---, storeNumber=873, registerID=0, startTime=2022-01-26T13:41:57, pin=null)]) | Headers-> ConsumerApp: STF & MessageId: 597c790b-6482-4a23-905c-49af6959241e & EndpointUrl: /v1/cash/getBalanceBulk & CorrelationId: null`),
		Entry("should recognize a complex INFO message", "info", `2025-06-16 10:32:54,862 INFO [http-nio-8080-exec-10] se.skv.chassi.MyResetController -- This is an example log message containing work ERROR `),
		Entry("should recognize a WARN not structured message with few key words (WARNING, Error): WARN met first ", "warn", `Apr 24 15:02:47 fedora gnome-session-binary[3056]: GnomeDesktop-WARNING: Could not create transient scope for PID 3303: GDBus.Error:org.freedesktop.DBus.Error.UnixProcessIdUnknown: Failed to set unit properties: No such process`),
	)
})

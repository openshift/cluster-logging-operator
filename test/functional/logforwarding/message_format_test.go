package logforwarding

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
)

var _ = Describe("[LogForwarding] Functional tests for message format", func() {

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should parse application log format correctly", func() {

		// Log message data
		message := "Functional test message"
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		time, _ := time.Parse(time.RFC3339Nano, timestamp)

		// Template expected as output Log
		var outputLogTemplate = types.ApplicationLog{
			Timestamp:     time,
			Message:       fmt.Sprintf("regex:^%s.*$", message),
			ViaqIndexName: "app-write",
			Level:         "unknown",
			ViaqMsgID:     "*",
			PipelineMetadata: types.PipelineMetadata{Collector: types.Collector{
				Ipaddr4:    "*",
				Inputname:  "*",
				Name:       "*",
				ReceivedAt: time,
				Version:    "*",
			},
			},
			Docker: types.Docker{
				ContainerID: "*"},
			Kubernetes: types.Kubernetes{
				ContainerName:     "*",
				PodName:           "*",
				NamespaceName:     "*",
				NamespaceID:       "*",
				OrphanedNamespace: "*"},
		}

		// Write log line as input to fluentd
		applicationLogLine := fmt.Sprintf("%s stdout F %s $n", timestamp, message)
		Expect(framework.WritesMessageToApplicationLogs(applicationLogLine, 1)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom("fluentforward")
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(raw, &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	})
})

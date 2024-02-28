package http

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

var _ = Describe("[Functional][Outputs][Http] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToHttpOutput()
		framework.Forwarder.Spec.Outputs[0].Tuning = &logging.OutputTuningSpec{
			Delivery:         logging.OutputDeliveryModeAtLeastOnce,
			MaxRetryDuration: utils.GetPtr(time.Duration(30)),
			MinRetryDuration: utils.GetPtr(time.Duration(5)),
			MaxWrite:         utils.GetPtr(resource.MustParse("1M")),
		}
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("", func(addDestinationContainer func(f *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor) {

		Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
			addDestinationContainer(framework),
			func(builder *runtime.PodBuilder) error {
				builder.AddLabels(map[string]string{
					"app.kubernetes.io/name": "somevalue",
					"foo.bar":                "a123",
				})
				return nil
			},
		})).To(BeNil())

		message := "hello world"
		timestamp := "2020-11-04T18:13:59.061892+00:00"

		applicationLogLine := fmt.Sprintf("%s stdout F %s", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())
		// Read line from Destination Http output
		result, err := framework.ReadFileFrom("http", functional.ApplicationLogFile)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(result).ToNot(BeEmpty())
		raw := strings.Split(strings.TrimSpace(result), "\n")
		logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
		Expect(err).To(BeNil(), fmt.Sprintf("Expected no errors parsing the logs: %s", raw[0]))
		// Compare to expected template
		Expect(logs[0].Message).To(Equal(message))
		Expect(logs[0].Kubernetes.Labels).To(HaveKey(MatchRegexp("^([a-zA-Z0-9_]*)$")))
		Expect(logs[0].Kubernetes.Labels).To(HaveKey(MatchRegexp("foo")))
	},
		Entry("should send message over http to vector", func(f *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddVectorHttpOutput(b, f.Forwarder.Spec.Outputs[0])
			}
		}),
		Entry("should send message over http to fluentd", func(f *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddFluentdHttpOutput(b, f.Forwarder.Spec.Outputs[0])
			}
		}),
	)

})

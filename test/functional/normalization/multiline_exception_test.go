package normalization

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"strings"
)

// Multiline Detect Exception test to verify proper re-assembly of
// multi-line exceptions (e.g. java stacktrace)
// https://issues.redhat.com/browse/LOG-1717
var _ = Describe("[Functional][Normalization] Multi-line exception detection", func() {
	const (
		timestamp = "2021-03-31T12:59:28.573159188+00:00"
	)

	var (
		javaException = `java.lang.NullPointerException: Cannot invoke "String.toString()" because "<parameter1>" is null
        at testjava.Main.printMe(Main.java:19)
        at testjava.Main.main(Main.java:10)`

		nodeJSException = `ReferenceError: myArray is not defined
  at next (/app/node_modules/express/lib/router/index.js:256:14)
  at /app/node_modules/express/lib/router/index.js:615:15
  at next (/app/node_modules/express/lib/router/index.js:271:10)
  at Function.process_params (/app/node_modules/express/lib/router/index.js:330:12)
  at /app/node_modules/express/lib/router/index.js:277:22
  at Layer.handle [as handle_request] (/app/node_modules/express/lib/router/layer.js:95:5)
  at Route.dispatch (/app/node_modules/express/lib/router/route.js:112:3)
  at next (/app/node_modules/express/lib/router/route.js:131:13)
  at Layer.handle [as handle_request] (/app/node_modules/express/lib/router/layer.js:95:5)
  at /app/app.js:52:3`

		goLangException = `panic: my panic

goroutine 4 [running]:
panic(0x45cb40, 0x47ad70)
	/usr/local/go/src/runtime/panic.go:542 +0x46c fp=0xc42003f7b8 sp=0xc42003f710 pc=0x422f7c
main.main.func1(0xc420024120)
	foo.go:6 +0x39 fp=0xc42003f7d8 sp=0xc42003f7b8 pc=0x451339
runtime.goexit()
	/usr/local/go/src/runtime/asm_amd64.s:2337 +0x1 fp=0xc42003f7e0 sp=0xc42003f7d8 pc=0x44b4d1
created by main.main
	foo.go:5 +0x58`
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		b := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			WithMultineErrorDetection().
			ToFluentForwardOutput()
		//LOG-2241
		b.FromInput(logging.InputNameApplication).
			Named("other").
			WithMultineErrorDetection().
			ToOutputWithVisitor(func(spec *logging.OutputSpec) {
				spec.Type = logging.OutputTypeFluentdForward
				spec.URL = "tcp://0.0.0.0:24234"
			}, "missing")
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("should reassemble multi-line stacktraces", func(exception string) {
		buffer := []string{}
		for _, line := range strings.Split(exception, "\n") {
			crioLine := functional.NewCRIOLogMessage(timestamp, line, false)
			buffer = append(buffer, crioLine)
		}
		Expect(framework.WriteMessagesToApplicationLog(strings.Join(buffer, "\n"), 1)).To(Succeed())
		raw, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs[0].Message).To(Equal(exception))
	},
		Entry("of Java services", javaException),
		Entry("of NodeJS services", nodeJSException),
		Entry("of GoLang services", goLangException),
	)

})

//go:build fluentd
// +build fluentd

package normalization

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
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

		jsClientSideException = `Error  
				at bls (<anonymous>:3:9)
				at <anonymous>:6:4
				at a_function_name
				at Object.InjectedScript._evaluateOn (https://<anonymous>/file.js?foo=bar:875:140)
				at Object.InjectedScript.evaluate (<anonymous>)`

		jsV8Exception = `V8 errors stack trace
			eval at Foo.a (eval at Bar.z (myscript.js:10:3))
			at new Contructor.Name (native)
			at new FunctionName (unknown location)
			at Type.functionName [as methodName] (file(copy).js?query='yes':12:9)
			at functionName [as methodName] (native)
			at Type.main(sample(copy).js:6:4)`

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

		appNamespace = "multi-line-test"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("should reassemble multi-line stacktraces", func(exception string, buildLogForwarder func(framework *functional.CollectorFunctionalFramework)) {

		if buildLogForwarder == nil {
			buildLogForwarder = func(framework *functional.CollectorFunctionalFramework) {
				functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					WithMultineErrorDetection().
					ToFluentForwardOutput()
			}
		}
		buildLogForwarder(framework)
		framework.VisitConfig = func(conf string) string {
			conf = strings.Replace(conf, "@type kubernetes_metadata", "@type kubernetes_metadata\ntest_api_adapter  KubernetesMetadata::TestApiAdapter\n", 1)
			return conf
		}
		Expect(framework.Deploy()).To(BeNil())

		buffer := []string{}
		for _, line := range strings.Split(exception, "\n") {
			crioLine := functional.NewCRIOLogMessage(timestamp, line, false)
			buffer = append(buffer, crioLine)
		}
		// Application log in namespace
		Expect(framework.WriteMessagesToNamespace(strings.Join(buffer, "\n"), appNamespace, 1)).To(Succeed())

		for _, output := range framework.Forwarder.Spec.Outputs {
			outputType := output.Type
			raw, err := framework.ReadRawApplicationLogsFrom(outputType)
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type %s", outputType)
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs for type %s: %s", outputType, raw)
			Expect(logs[0].Message).To(Equal(exception))
		}
	},
		Entry("of Java services", javaException, nil),
		Entry("of JS client side exception", jsClientSideException, nil),
		Entry("of V8 errors stack trace", jsV8Exception, nil),
		Entry("of NodeJS services", nodeJSException, nil),
		Entry("of GoLang services", goLangException, nil),
		Entry("of single application NS to single pipeline", goLangException, func(framework *functional.CollectorFunctionalFramework) {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("forward-pipeline", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace},
					}
				}).
				WithMultineErrorDetection().
				ToFluentForwardOutput()
		}),
		Entry("of single application NS sources with multiple pipelines", goLangException, func(framework *functional.CollectorFunctionalFramework) {
			b := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("multiline-log-ns", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace},
					}
				}).
				WithMultineErrorDetection().
				ToFluentForwardOutput()
			//LOG-2241
			b.FromInput("multiline-log-ns").
				Named("other").
				WithMultineErrorDetection().
				ToElasticSearchOutput()
		}),
		Entry("of multiple application NS source with multiple pipelines", goLangException, func(framework *functional.CollectorFunctionalFramework) {
			b := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("multiline-log-ns", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace, "multi-line-test-2"},
					}
				}).
				WithMultineErrorDetection().
				ToFluentForwardOutput()
			//LOG-2241
			b.FromInput("multiline-log-ns").
				Named("other").
				WithMultineErrorDetection().
				ToElasticSearchOutput()
		}),
	)

})

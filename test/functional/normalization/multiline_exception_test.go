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
// https://issues.redhat.com/browse/LOG-1796
var _ = Describe("[Functional][Normalization] Multi-line exception detection", func() {
	const (
		timestamp = "2021-03-31T12:59:28.573159188+00:00"
	)

	var (
		javaException = `java.lang.NullPointerException: Cannot invoke "String.toString()" because "<parameter1>" is null
        at testjava.Main.printMe(Main.java:19)
        at testjava.Main.main(Main.java:10)`

		javaExceptionComp = `java.lang.IndexOutOfBoundsException: Index: 1, Size: 0
		at java.util.ArrayList.rangeCheck(ArrayList.java:657)
		at java.util.ArrayList.get(ArrayList.java:433)
		at com.in28minutes.rest.webservices.restfulwebservices.HelloWorldController.helloWorld(HelloWorldController.java:14)
		at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
		at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
		at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
		at java.lang.reflect.Method.invoke(Method.java:498)
		at org.springframework.web.method.support.InvocableHandlerMethod.doInvoke(InvocableHandlerMethod.java:190)
		at org.springframework.web.method.support.InvocableHandlerMethod.invokeForRequest(InvocableHandlerMethod.java:138)
		at org.springframework.web.servlet.mvc.method.annotation.ServletInvocableHandlerMethod.invokeAndHandle(ServletInvocableHandlerMethod.java:104)
		at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.invokeHandlerMethod(RequestMappingHandlerAdapter.java:892)
		at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.handleInternal(RequestMappingHandlerAdapter.java:797)
		at org.springframework.web.servlet.mvc.method.AbstractHandlerMethodAdapter.handle(AbstractHandlerMethodAdapter.java:87)
		at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1039)
		at org.springframework.web.servlet.DispatcherServlet.doService(DispatcherServlet.java:942)
		at org.springframework.web.servlet.FrameworkServlet.processRequest(FrameworkServlet.java:1005)
		at org.springframework.web.servlet.FrameworkServlet.doGet(FrameworkServlet.java:897)
		at javax.servlet.http.HttpServlet.service(HttpServlet.java:634)
		at org.springframework.web.servlet.FrameworkServlet.service(FrameworkServlet.java:882)
		at javax.servlet.http.HttpServlet.service(HttpServlet.java:741)
		at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:231)
		at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
		at org.apache.tomcat.websocket.server.WsFilter.doFilter(WsFilter.java:53)
		at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:193)
		at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
		at org.springframework.web.filter.RequestContextFilter.doFilterInternal(RequestContextFilter.java:99)
		at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:118)
		at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:193)
		at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
		at org.springframework.web.filter.FormContentFilter.doFilterInternal(FormContentFilter.java:92)
		at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:118)
		at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:193)
		at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
		at org.springframework.web.filter.HiddenHttpMethodFilter.doFilterInternal(HiddenHttpMethodFilter.java:93)
		at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:118)
		at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:193)
		at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
		at org.springframework.web.filter.CharacterEncodingFilter.doFilterInternal(CharacterEncodingFilter.java:200)
		at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:118)
		at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:193)
		at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
		at org.apache.catalina.core.StandardWrapperValve.invoke(StandardWrapperValve.java:202)
		at org.apache.catalina.core.StandardContextValve.invoke(StandardContextValve.java:96)
		at org.apache.catalina.authenticator.AuthenticatorBase.invoke(AuthenticatorBase.java:490)
		at org.apache.catalina.core.StandardHostValve.invoke(StandardHostValve.java:139)
		at org.apache.catalina.valves.ErrorReportValve.invoke(ErrorReportValve.java:92)
		at org.apache.catalina.core.StandardEngineValve.invoke(StandardEngineValve.java:74)
		at org.apache.catalina.connector.CoyoteAdapter.service(CoyoteAdapter.java:343)
		at org.apache.coyote.http11.Http11Processor.service(Http11Processor.java:408)
		at org.apache.coyote.AbstractProcessorLight.process(AbstractProcessorLight.java:66)
		at org.apache.coyote.AbstractProtocol$ConnectionHandler.process(AbstractProtocol.java:853)
		at org.apache.tomcat.util.net.NioEndpoint$SocketProcessor.doRun(NioEndpoint.java:1587)
		at org.apache.tomcat.util.net.SocketProcessorBase.run(SocketProcessorBase.java:49)
		at java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1149)
		at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:624)
		at org.apache.tomcat.util.threads.TaskThread$WrappingRunnable.run(TaskThread.java:61)
		at java.lang.Thread.run(Thread.java:748)
	Caused by: com.example.myproject.MyProjectServletException
		at com.example.myproject.MyServlet.doPost(MyServlet.java:169)
		at javax.servlet.http.HttpServlet.service(HttpServlet.java:727)
		at javax.servlet.http.HttpServlet.service(HttpServlet.java:820)
		at org.mortbay.jetty.servlet.ServletHolder.handle(ServletHolder.java:511)
		at org.mortbay.jetty.servlet.ServletHandler$CachedChain.doFilter(ServletHandler.java:1166)
		at com.example.myproject.OpenSessionInViewFilter.doFilter(OpenSessionInViewFilter.java:30)
		... 27 common frames omitted`

		pythonException = `Traceback (most recent call last):
  File "/usr/bin/python/python27/python27_lib/versions/third_party/webapp2-2.5.2/webapp2.py", line 1535, in __call__
    rv = self.handle_exception(request, response, e)
  File "/home/apps/myapp/app.py", line 10, in start
    return get()
  File "/home/apps/myapp/app.py", line 5, in get
    raise Exception('exception')
Exception: ('exception')`

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
					ToElasticSearchOutput()
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

		if testfw.LogCollectionType == logging.LogCollectionTypeVector {
			Expect(framework.WriteMessagesToApplicationLog(strings.Join(buffer, "\n"), 1)).To(Succeed())
		} else {
			Expect(framework.WriteMessagesToNamespace(strings.Join(buffer, "\n"), appNamespace, 1)).To(Succeed())
		}

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
		Entry("of Java services more advance", javaExceptionComp, nil),
		Entry("of NodeJS services", nodeJSException, nil),
		Entry("of GoLang services", goLangException, nil),
		Entry("of Python services", pythonException, nil),
		Entry("of single application NS to single pipeline", goLangException, func(framework *functional.CollectorFunctionalFramework) {
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("not a valid test for vector since we route by namespace")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("forward-pipeline", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace},
					}
				}).
				WithMultineErrorDetection().
				ToElasticSearchOutput()
		}),
		Entry("of single application NS sources with multiple pipelines", goLangException, func(framework *functional.CollectorFunctionalFramework) {
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("not a valid test for vector since we route by namespace")
			}
			b := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("multiline-log-ns", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace},
					}
				}).
				WithMultineErrorDetection().
				ToElasticSearchOutput()
			//LOG-2241
			b.FromInput("multiline-log-ns").
				Named("other").
				WithMultineErrorDetection().
				ToElasticSearchOutput()
		}),
		Entry("of multiple application NS source with multiple pipelines", goLangException, func(framework *functional.CollectorFunctionalFramework) {
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("not a valid test for vector since we route by namespace")
			}
			b := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("multiline-log-ns", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace, "multi-line-test-2"},
					}
				}).
				WithMultineErrorDetection().
				ToElasticSearchOutput()
			//LOG-2241
			b.FromInput("multiline-log-ns").
				Named("other").
				WithMultineErrorDetection().
				ToElasticSearchOutput()
		}),
	)

})

//go:build fluentd
// +build fluentd

package normalization

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/matchers"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

// Fast test for checking reassembly logic for split log by CRI-O.
// CRI-O split long string on parts. Part marked by 'P' letter and finished with 'F'.
// Example:
// 2021-03-31T12:59:28.573159188+00:00 stdout P First line of log entry
// 2021-03-31T12:59:28.573159188+00:00 stdout P Second line of the log entry
// 2021-03-31T12:59:28.573159188+00:00 stdout F Last line of the log entry
//
// Here we will emulate CRI-O split by direct writing formatted content
var _ = Describe("[Functional][Normalization]Reassembly split by CRI-O logs ", func() {

	const chunkSize = 1024 * 8
	var (
		framework *functional.CollectorFunctionalFramework
		timestamp = "2021-03-31T12:59:28.573159188+00:00"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should handle a single split log message", func() {
		//write partial log
		msg := functional.NewCRIOLogMessage(timestamp, "May ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write partial log
		msg = functional.NewCRIOLogMessage(timestamp, "the force ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write partial log
		msg = functional.NewCRIOLogMessage(timestamp, "be with ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write final part of log entry
		msg = functional.NewCRIOLogMessage(timestamp, "you", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs[0].Message).Should(Equal("May the force be with you"))
	})

	It("should handle a split log followed by a full log", func() {
		//write partial log
		msg := functional.NewCRIOLogMessage(timestamp, "Run, ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write partial log
		msg = functional.NewCRIOLogMessage(timestamp, "Forest, ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write final part of log entry
		msg = functional.NewCRIOLogMessage(timestamp, "Run!", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write single-line log entry
		msg = functional.NewCRIOLogMessage(timestamp, "Freedom!!!", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")

		logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(logs)).To(Equal(2))
		Expect(logs[0].Message).Should(Equal("Run, Forest, Run!"))
		Expect(logs[1].Message).Should(Equal("Freedom!!!"))
	})

	Context("When a Java container logs a multi-line stack trace as JSON", func() {
		It("should be forwarded as a single message", func() {
			stack := strings.ReplaceAll(JsonJavaStackTrace, "\n", "")
			messages := []string{}
			start := 0
			totChunks := (len(stack) / (chunkSize)) + 1
			partial := true
			for chunk := 0; chunk < totChunks; chunk++ {
				end := start + chunkSize
				if end > len(stack) {
					end = len(stack)
					partial = false
				}
				m := stack[start:end]
				messages = append(messages, functional.NewCRIOLogMessage(timestamp, m, partial))
				start = end
			}
			message := strings.Join(messages, "\n")
			Expect(framework.WriteMessagesToApplicationLog(message, 1)).To(Succeed())
			logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(logs[0].Message).Should(Equal(stack))
		})
	})
})

// JsonJavaStackTrace - sample of Java stack trace in JSON format
const JsonJavaStackTrace = `{"instant":{"epochSecond":1617711572,"nanoOfSecond":619500000},"thread":"main",
"level":"ERROR","loggerName":"App","message":"trace","thrown":{"commonElementCount":0,
"localizedMessage":"/ by zero","message":"/ by zero","name":"java.lang.ArithmeticException",
"extendedStackTrace":[{"class":"App","method":"main","file":"App.java","line":15,
"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?",
"version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke",
"file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method",
"method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},
{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java",
"line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main",
"file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},{"class":
"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?",
"version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke",
"file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,
"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main",
"file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?",
"version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke",
"file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,
"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main",
"file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},
{"class":"App","method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},
{"class":"App","method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,
"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke",
"file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},{"class":
"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,
"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java",
"line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main",
"file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main",
"file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,
"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},
{"class":"App","method":"main","file":"App.java","line":15,"exact":true,"location":"classes/",
"version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke",
"file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},
{"class":"App","method":"main","file":"App.java","line":15,"exact":true,"location":"classes/",
"version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},
{"class":"App","method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,
"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke",
"file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,
"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke",
"file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,
"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main",
"file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App",
"method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java",
"line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke","file":"DelegatingMethodAccessorImpl.java",
"line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method","method":"invoke",
"file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},{"class":
"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java","line":131,
"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java",
"line":15,"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java",
"line":62,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?",
"version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java",
"line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java",
"line":15,"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java",
"line":62,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?",
"version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java","line":131,
"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java","line":15,"exact":true,
"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0",
"file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},{"class":
"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,
"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?",
"version":"?"},{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,
"exact":false,"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2",
"method":"main","file":"AppMainV2.java","line":131,"exact":true,"location":"idea_rt.jar","version":"?"},
{"class":"App","method":"main","file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,
"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java",
"line":62,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?",
"version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java",
"line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java",
"line":15,"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,
"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke",
"file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method",
"method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},
{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java",
"line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main",
"file":"App.java","line":15,"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl",
"method":"invoke0","file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java","line":62,
"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl","method":"invoke",
"file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},{"class":"java.lang.reflect.Method",
"method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},
{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java","line":131,
"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java",
"line":15,"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0",
"file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke",
"file":"NativeMethodAccessorImpl.java","line":62,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?",
"version":"?"},{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,
"location":"?","version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java",
"line":131,"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java","line":15,
"exact":true,"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0",
"file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},{"class":
"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java",
"line":62,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?",
"version":"?"},{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java","line":131,
"exact":true,"location":"idea_rt.jar","version":"?"},{"class":"App","method":"main","file":"App.java","line":15,"exact":true,
"location":"classes/","version":"?"},{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke0",
"file":"NativeMethodAccessorImpl.java","line":-2,"exact":false,"location":"?","version":"?"},
{"class":"jdk.internal.reflect.NativeMethodAccessorImpl","method":"invoke","file":"NativeMethodAccessorImpl.java",
"line":62,"exact":false,"location":"?","version":"?"},{"class":"jdk.internal.reflect.DelegatingMethodAccessorImpl",
"method":"invoke","file":"DelegatingMethodAccessorImpl.java","line":43,"exact":false,"location":"?","version":"?"},
{"class":"java.lang.reflect.Method","method":"invoke","file":"Method.java","line":566,"exact":false,"location":"?","version":"?"},
{"class":"com.intellij.rt.execution.application.AppMainV2","method":"main","file":"AppMainV2.java","line":131,"exact":true,
"location":"idea_rt.jar","version":"?"}]},"endOfBatch":false,"loggerFqcn":"org.apache.logging.log4j.spi.AbstractLogger",
"threadId":1,"threadPriority":5,"service":"java-app"}`

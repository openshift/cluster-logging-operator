package syslog

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/format"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"strings"
	"time"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var _ = Describe("[Functional][Outputs][Syslog] RFC5424 tests", func() {

	var (
		framework          *functional.CollectorFunctionalFramework
		maxReadDuration, _ = time.ParseDuration("30s")
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		framework.MaxReadDuration = &maxReadDuration
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("logforwarder configured with appname, msgid, and procid", func(appName, msgId, procId, payloadKey, expInfo string) {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			WithParseJson().
			ToSyslogOutput(obs.SyslogRFC5424, func(output *obs.OutputSpec) {
				output.Syslog.Facility = "user"
				output.Syslog.Severity = "debug"
				output.Syslog.AppName = appName
				output.Syslog.ProcId = procId
				output.Syslog.MsgId = msgId
				output.Syslog.PayloadKey = payloadKey
			})
		Expect(framework.Deploy()).To(BeNil())
		format.MaxLength = 0
		record := `{"index":1,"appname_key":"rec_appname","msgid_key":"rec_msgid","procid_key":"rec_procid"}`
		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		expMatch := fmt.Sprintf(`( %s )`, expInfo)
		Expect(outputlogs[0]).To(MatchRegexp(expMatch), "Exp to match the appname/procid/msgid in received message")
		jsonPart := outputlogs[0][strings.Index(outputlogs[0], "{"):]
		logs, err := types.ParseLogs(utils.ToJsonLogs([]string{jsonPart}))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs).To(HaveLen(1), "Expected the receiver to receive the message")
		var expected map[string]interface{}
		err = json.Unmarshal([]byte(record), &expected)
		Expect(err).To(BeNil(), "Expected no errors unmarshaling the expected json")
		same := cmp.Equal(logs[0].Structured, expected)
		if !same {
			diff := cmp.Diff(logs[0].Structured, expected)
			fmt.Printf("diff %s\n", diff)
		}
		Expect(same).To(BeTrue(), "parsed json message not matching")
	},
		Entry("should use the value from the record and include the structured message", `{.structured.appname_key||"none"}`, `{.structured.msgid_key||"none"}`, `{.structured.procid_key||"none"}`, `{.message}`, "rec_appname rec_procid rec_msgid"),
	)

	DescribeTable("logforwarder configured with payload key", func(appName, msgId, procId, payloadKey string) {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			WithParseJson().
			ToSyslogOutput(obs.SyslogRFC5424, func(output *obs.OutputSpec) {
				output.Syslog.Facility = "user"
				output.Syslog.Severity = "debug"
				output.Syslog.AppName = appName
				output.Syslog.ProcId = procId
				output.Syslog.MsgId = msgId
				output.Syslog.PayloadKey = payloadKey
			})
		Expect(framework.Deploy()).To(BeNil())

		record := `{"index":1,"appname_key":"rec_appname","msgid_key":"rec_msgid","procid_key":"rec_procid"}`
		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		fields := strings.Split(outputlogs[0], " - ")
		payload := strings.TrimSpace(fields[1])
		parsedRecord := map[string]interface{}{}
		Expect(json.Unmarshal([]byte(payload), &parsedRecord)).To(BeNil(), fmt.Sprintf("payload: %q", payload))
		msg := parsedRecord["structured"]
		var expected map[string]interface{}
		err = json.Unmarshal([]byte(record), &expected)
		Expect(err).To(BeNil(), "Expected no errors unmarshaling the expected json")
		same := cmp.Equal(msg, expected)
		if !same {
			diff := cmp.Diff(msg, expected)
			fmt.Printf("diff %s\n", diff)
		}
		Expect(same).To(BeTrue(), "parsed json message not matching")
	},
		Entry("should include the structured message when payloadkey is empty", `{.structured.appname_key||"none"}`, `{.structured.msgid_key||"none"}`, `{.structured.procid_key||"none"}`, ""),
		Entry("should include the message when payloadkey is not found", `{.structured.appname_key||"none"}`, `{.structured.msgid_key||"none"}`, `{.structured.procid_key||"none"}`, "{.key_not_available}"),
	)

	Describe("configured with values for facility,severity", func() {
		It("should use values from the record", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithParseJson().
				ToSyslogOutput(obs.SyslogRFC5424, func(spec *obs.OutputSpec) {
					spec.Syslog.Facility = `{.structured.facility_key||"notfound"}`
					spec.Syslog.Severity = `{.structured.severity_key||"notfound"}`
				})
			Expect(framework.Deploy()).To(BeNil())

			record := `{"index":1,"timestamp":1,"facility_key":"local0","severity_key":"Informational"}`
			crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
			Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")

			// 134 = Facility(local0/16)*8 + Severity(Informational/6)
			// The 1 after <134> is version, which is always set to 1
			expectedPriority := "<134>1 "
			Expect(outputlogs[0]).To(MatchRegexp(expectedPriority), "Exp to find tag in received message")
		})

		It("should use numeric value", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, func(spec *obs.OutputSpec) {
					spec.Syslog.Facility = "16"
					spec.Syslog.Severity = "6"
				})
			Expect(framework.Deploy()).To(BeNil())

			record := `{"index":1,"timestamp":1}`
			crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
			Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")

			// 134 = Facility(local0/16)*8 + Severity(Informational/6)
			// The 1 after <134> is version, which is always set to 1
			expectedPriority := "<134>1 "
			Expect(outputlogs[0]).To(MatchRegexp(expectedPriority), "Exp to find tag in received message")
		})
	})
	It("should be able to send a large payload", func() {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToSyslogOutput(obs.SyslogRFC5424)
		Expect(framework.Deploy()).To(BeNil())

		record := strings.ReplaceAll(largeStackTrace, "\n", "    ")
		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		Expect(outputlogs[0]).To(MatchRegexp(`java\.lang.*GroovyStarter.*131`), "Exp to find tag in received message")
	})

})

const (
	largeStackTrace = `java.lang.RuntimeException: Error grabbing Grapes -- [unresolved dependency: org.jenkins-ci#trilead-ssh2;build-217-jenkins-14: not found, unresolved dependency: org.kohsuke.stapler#stapler-groovy;1.256: not found, unresolved dependency: org.kohsuke.stapler#stapler-jrebel;1.256: not found, unresolved dependency: io.jenkins.stapler#jenkins-stapler-support;1.0: not found, unresolved dependency: commons-httpclient#commons-httpclient;3.1-jenkins-1: not found, unresolved dependency: org.jenkins-ci#bytecode-compatibility-transformer;2.0-beta-2: not found, unresolved dependency: org.jenkins-ci#task-reactor;1.5: not found, download failed: org.jenkins-ci.main#jenkins-core;2.164!jenkins-core.jar, download failed: org.jenkins-ci.plugins.icon-shim#icon-set;1.0.5!icon-set.jar, download failed: org.jenkins-ci.main#remoting;3.29!remoting.jar, download failed: org.jenkins-ci#constant-pool-scanner;1.2!constant-pool-scanner.jar, download failed: org.jenkins-ci.main#cli;2.164!cli.jar, download failed: org.kohsuke#access-modifier-annotation;1.14!access-modifier-annotation.jar, download failed: org.jenkins-ci#annotation-indexer;1.12!annotation-indexer.jar, download failed: org.jvnet.localizer#localizer;1.24!localizer.jar, download failed: net.i2p.crypto#eddsa;0.3.0!eddsa.jar(bundle), download failed: org.jenkins-ci#version-number;1.6!version-number.jar, download failed: com.google.code.findbugs#annotations;3.0.0!annotations.jar, download failed: org.jenkins-ci#crypto-util;1.1!crypto-util.jar, download failed: org.jvnet.hudson#jtidy;4aug2000r7-dev-hudson-1!jtidy.jar, download failed: com.google.inject#guice;4.0!guice.jar, download failed: org.jruby.ext.posix#jna-posix;1.0.3-jenkins-1!jna-posix.jar, download failed: com.github.jnr#jnr-posix;3.0.45!jnr-posix.jar, download failed: com.github.jnr#jnr-ffi;2.1.8!jnr-ffi.jar, download failed: org.slf4j#jcl-over-slf4j;1.7.25!jcl-over-slf4j.jar, download failed: org.slf4j#log4j-over-slf4j;1.7.25!log4j-over-slf4j.jar, download failed: javax.xml.stream#stax-api;1.0-2!stax-api.jar]
	at sun.reflect.NativeConstructorAccessorImpl.newInstance0(Native Method)
	at sun.reflect.NativeConstructorAccessorImpl.newInstance(NativeConstructorAccessorImpl.java:62)
	at sun.reflect.DelegatingConstructorAccessorImpl.newInstance(DelegatingConstructorAccessorImpl.java:45)
	at java.lang.reflect.Constructor.newInstance(Constructor.java:423)
	at org.codehaus.groovy.reflection.CachedConstructor.invoke(CachedConstructor.java:83)
	at org.codehaus.groovy.reflection.CachedConstructor.doConstructorInvoke(CachedConstructor.java:77)
	at org.codehaus.groovy.runtime.callsite.ConstructorSite$ConstructorSiteNoUnwrap.callConstructor(ConstructorSite.java:84)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallConstructor(CallSiteArray.java:60)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callConstructor(AbstractCallSite.java:235)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callConstructor(AbstractCallSite.java:247)
	at groovy.grape.GrapeIvy.getDependencies(GrapeIvy.groovy:424)
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.codehaus.groovy.runtime.callsite.PogoMetaMethodSite$PogoCachedMethodSite.invoke(PogoMetaMethodSite.java:169)
	at org.codehaus.groovy.runtime.callsite.PogoMetaMethodSite.callCurrent(PogoMetaMethodSite.java:59)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallCurrent(CallSiteArray.java:52)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:154)
	at groovy.grape.GrapeIvy.resolve(GrapeIvy.groovy:571)
	at groovy.grape.GrapeIvy$resolve$1.callCurrent(Unknown Source)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallCurrent(CallSiteArray.java:52)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:154)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:190)
	at groovy.grape.GrapeIvy.resolve(GrapeIvy.groovy:538)
	at groovy.grape.GrapeIvy$resolve$0.callCurrent(Unknown Source)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallCurrent(CallSiteArray.java:52)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:154)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:182)
	at groovy.grape.GrapeIvy.grab(GrapeIvy.groovy:256)
	at groovy.grape.Grape.grab(Grape.java:167)
	at groovy.grape.GrabAnnotationTransformation.visit(GrabAnnotationTransformation.java:378)
	at org.codehaus.groovy.transform.ASTTransformationVisitor$3.call(ASTTransformationVisitor.java:321)
	at org.codehaus.groovy.control.CompilationUnit.applyToSourceUnits(CompilationUnit.java:943)
	at org.codehaus.groovy.control.CompilationUnit.doPhaseOperation(CompilationUnit.java:605)
	at org.codehaus.groovy.control.CompilationUnit.processPhaseOperations(CompilationUnit.java:581)
	at org.codehaus.groovy.control.CompilationUnit.compile(CompilationUnit.java:558)
	at groovy.lang.GroovyClassLoader.doParseClass(GroovyClassLoader.java:298)
	at groovy.lang.GroovyClassLoader.parseClass(GroovyClassLoader.java:268)
	at groovy.lang.GroovyShell.parseClass(GroovyShell.java:688)
	at groovy.lang.GroovyShell.run(GroovyShell.java:517)
	at groovy.lang.GroovyShell.run(GroovyShell.java:507)
	at groovy.ui.GroovyMain.processOnce(GroovyMain.java:653)
	at groovy.ui.GroovyMain.run(GroovyMain.java:384)
	at groovy.ui.GroovyMain.process(GroovyMain.java:370)
	at groovy.ui.GroovyMain.processArgs(GroovyMain.java:129)
	at groovy.ui.GroovyMain.main(GroovyMain.java:109)
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.codehaus.groovy.tools.GroovyStarter.rootLoader(GroovyStarter.java:109)
	at org.codehaus.groovy.tools.GroovyStarter.main(GroovyStarter.java:131)
	`
)

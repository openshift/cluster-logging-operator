package syslog

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
)

var _ = Describe("[Functional][Outputs][Syslog] Functional tests", func() {
	const (
		PRI_14  = "<14>1"
		PRI_15  = "<15>1"
		PRI_110 = "<110>1"
	)

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)

	})
	AfterEach(func() {
		framework.Cleanup()
	})

	setSyslogSpecValues := func(outspec *logging.OutputSpec) {
		outspec.Syslog = &logging.Syslog{
			Facility: "user",
			Severity: "debug",
			AppName:  "myapp",
			ProcID:   "myproc",
			MsgID:    "mymsg",
			RFC:      "RFC5424",
		}
	}

	join := func(
		f1 func(spec *logging.OutputSpec),
		f2 func(spec *logging.OutputSpec)) func(*logging.OutputSpec) {
		return func(s *logging.OutputSpec) {
			f1(s)
			f2(s)
		}
	}
	getTag := func(log string) string {
		for strings.Contains(log, "  ") {
			log = strings.ReplaceAll(log, "  ", " ")
		}
		fields := strings.Split(log, " ")
		return strings.TrimSuffix(fields[4], ":")
	}

	getPri := func(fields []string) string {
		return fields[0]
	}
	getAppName := func(fields []string) string {
		return fields[3]
	}
	getProcID := func(fields []string) string {
		return fields[4]
	}
	getMsgID := func(fields []string) string {
		return fields[5]
	}

	timestamp := "2013-03-28T14:36:03.243000+00:00"

	Context("Application Logs", func() {
		It("RFC5424: should send large message over UDP", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setSyslogSpecValues, func(spec *logging.OutputSpec) {
					spec.URL = "udp://0.0.0.0:24224"
				}), logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			var MaxLen int = 30000
			Expect(framework.WritesNApplicationLogsOfSize(1, MaxLen, 0)).To(BeNil())
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " - ")
			payload := strings.TrimSpace(fields[1])
			record := map[string]interface{}{}
			Expect(json.Unmarshal([]byte(payload), &record)).To(BeNil())
			msg := record["message"]
			var message string
			message, ok := msg.(string)
			Expect(ok).To(BeTrue())
			ReceivedLen := uint64(len(message))
			Expect(ReceivedLen).To(BeEquivalentTo(MaxLen), "Expected the message length to be the same")
		})

		It("should set appropriate syslog parameters containing `$`", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog = &logging.Syslog{
						RFC:     "RFC5424",
						AppName: "myapp$withdollar",
						MsgID:   "mymsgWith$dollar",
						ProcID:  "someProc$ID",
					}
					spec.URL = "udp://0.0.0.0:24224"
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(1, 1000, 0)).To(BeNil())
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp$withdollar"))
			Expect(getProcID(fields)).To(Equal("someProc$ID"))
			Expect(getMsgID(fields)).To(Equal("mymsgWith$dollar"))
		})

		It("RFC5424: should send NonJson App logs to syslog", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range NonJsonAppLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})

		It("RFC5424: should take values of appname, procid, messageid from record", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setSyslogSpecValues, func(spec *logging.OutputSpec) {
					spec.Syslog.AppName = "$.message.appname_key"
					spec.Syslog.ProcID = "$.message.procid_key"
					spec.Syslog.MsgID = "$.message.msgid_key"
				}), logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("rec_appname"))
			Expect(getProcID(fields)).To(Equal("rec_procid"))
			Expect(getMsgID(fields)).To(Equal("rec_msgid"))
		})
		It("should take values from fluent tag", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeFluentd {
				Skip("Test requires fluentd")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setSyslogSpecValues, func(spec *logging.OutputSpec) {
					spec.Syslog.AppName = "tag"
				}), logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = test.Escapelines(log)
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(HavePrefix("kubernetes."))
		})

		It("RFC3164: should take default value for appname", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = logging.SyslogRFC3164
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			app := strings.Join([]string{framework.Namespace, framework.Pod.Name, constants.CollectorName}, "")
			re := regexp.MustCompile("[^a-zA-Z0-9]")
			app = re.ReplaceAllString(app, "")
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			Expect(getTag(outputlogs[0])).To(Equal(app))
		})

		It("RFC3164: should take values of  tag from record", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.Tag = `$.message.appname_key`
					spec.Syslog.RFC = logging.SyslogRFC3164
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			Expect(getTag(outputlogs[0])).To(Equal("rec_appname"))
		})

	})
	Context("Audit logs", func() {
		It("RFC5424: should send kubernetes audit logs", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			Expect(framework.WriteK8sAuditLog(1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})
		It("RFC5424:should send openshift audit logs", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(Succeed())

			// Log message data
			Expect(framework.WriteOpenshiftAuditLog(1)).To(Succeed())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})

		It("RFC5424: should send kubernetes audit logs with default appname", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = string(logging.SyslogRFC5424)
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			Expect(framework.WriteK8sAuditLog(1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("kubeAPI"))
			Expect(getMsgID(fields)).To(Equal("kubeAPI"))
		})

		It("RFC5424: should send openshift audit logs with default appname", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = string(logging.SyslogRFC5424)
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			Expect(framework.WriteOpenshiftAuditLog(1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			fields := strings.Split(outputlogs[0], " ")
			Expect(getPri(fields)).To(Equal(PRI_110))
			Expect(getAppName(fields)).To(Equal("openshiftAPI"))
			Expect(getMsgID(fields)).To(Equal("openshiftAPI"))
		})

		It("RFC3164: should send kubernetes audit logs", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = logging.SyslogRFC3164
					spec.Syslog.Tag = "myapp"
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			Expect(framework.WriteK8sAuditLog(1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			Expect(getTag(outputlogs[0])).To(Equal("myapp"))
		})
		It("RFC3164:should send openshift audit logs", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = logging.SyslogRFC3164
					spec.Syslog.Tag = "myapp"
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(Succeed())

			// Log message data
			Expect(framework.WriteOpenshiftAuditLog(1)).To(Succeed())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			Expect(getTag(outputlogs[0])).To(Equal("myapp"))
		})

		It("RFC3164:should send openshift audit logs", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = logging.SyslogRFC3164
					spec.Syslog.Tag = "myapp"
					spec.Syslog.ProcID = "1243"
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(Succeed())

			// Log message data
			Expect(framework.WriteOpenshiftAuditLog(1)).To(Succeed())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			Expect(getTag(outputlogs[0])).To(Equal("myapp[1243]"))
		})

	})

	Context("Infrastructure log logs", func() {

		It("RFC5424: should send infra logs", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameInfrastructure).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			logline := functional.NewJournalLog(3, "*", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadInfrastructureLogsFrom(string(logging.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})

		It("RFC3164: should calc appname and pid to tag", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameInfrastructure).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = logging.SyslogRFC3164
					spec.Syslog.Tag = "myapp"
					spec.Syslog.ProcID = "1234"
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			logline := functional.NewJournalLog(3, "*", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadInfrastructureLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			Expect(getTag(outputlogs[0])).To(Equal("myapp[1234]"))
		})

		It("RFC5424: should take default values of appname, procid, messageid", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameInfrastructure).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = string(logging.SyslogRFC5424)
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			logline := functional.NewJournalLog(3, "*", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			// Read line from Syslog output

			outputlogs, err := framework.ReadInfrastructureLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("google-chrome.desktop"))
			Expect(getProcID(fields)).To(Equal("3194"))
			Expect(getMsgID(fields)).To(Equal("node"))
		})

		It("RFC3164: should take default values of tag", func() {
			if testfw.LogCollectionType != logging.LogCollectionTypeVector {
				Skip("Test requires Vector")
			}
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameInfrastructure).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.RFC = logging.SyslogRFC3164
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			logline := functional.NewJournalLog(3, "*", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			// Read line from Syslog output

			outputlogs, err := framework.ReadInfrastructureLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			Expect(getTag(outputlogs[0])).To(Equal("google-chrome.desktop[3194]"))
		})

	})

})

var (
	JSONApplicationLogs = []string{
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:52"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:53"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:54"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:55"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:56"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:57"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:58"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:59"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:55:00"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:55:01"}`,
	}

	NonJsonAppLogs = []string{
		`2021-02-17 17:46:27 "hello world"`,
		`2021-02-17 17:46:28 "hello world"`,
		`2021-02-17 17:46:29 "hello world"`,
		`2021-02-17 17:46:30 "hello world"`,
		`2021-02-17 17:46:31 "hello world"`,
		`2021-02-17 17:46:32 "hello world"`,
		`2021-02-17 17:46:33 "hello world"`,
		`2021-02-17 17:46:34 "hello world"`,
		`2021-02-17 17:46:35 "hello world"`,
		`2021-02-17 17:46:36 "hello world"`,
	}
)

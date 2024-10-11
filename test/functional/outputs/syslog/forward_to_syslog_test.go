package syslog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/common/secrets"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var _ = Describe("[Functional][OutputConditions][Syslog] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework

		setSyslogSpecValues = func(outspec *obs.OutputSpec) {
			outspec.Syslog.Facility = "user"
			outspec.Syslog.Severity = "debug"
			outspec.Syslog.AppName = "myapp"
			outspec.Syslog.ProcId = "myproc"
			outspec.Syslog.MsgId = "mymsg"
		}

		join = func(
			f1 func(spec *obs.OutputSpec),
			f2 func(spec *obs.OutputSpec)) func(*obs.OutputSpec) {
			return func(s *obs.OutputSpec) {
				f1(s)
				f2(s)
			}
		}

		getAppName = func(fields []string) string {
			return fields[3]
		}
		getProcID = func(fields []string) string {
			return fields[4]
		}
		getMsgID = func(fields []string) string {
			return fields[5]
		}

		timestamp = "2013-03-28T14:36:03.243000+00:00"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should be configurable to use TLS", func() {
		secret, tlsSpec := secrets.NewTLSSecret(framework.Forwarder.Namespace, "syslog-tls", framework.Namespace, "syslogreceiver")
		framework.Secrets = append(framework.Secrets, secret)
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToSyslogOutput(obs.SyslogRFC5424, func(output *obs.OutputSpec) {
				if output.TLS == nil {
					output.TLS = &obs.OutputTLSSpec{}
				}
				output.TLS.TLSSpec = tlsSpec
			})
		Expect(framework.Deploy()).To(BeNil())

		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), NonJsonAppLogs[0])
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(Not(BeEmpty()), "Expected the receiver to receive the message")
		Expect(outputlogs[0]).To(MatchRegexp(`^<[0-9]*>[0-9]* ([0-9T:.Z\-]*) ([a-z0-9-.]*) (.*) -(.*)-(.*)- (.*)$`), "Exp a syslog formatted message")
	})
	Context("Application Logs", func() {
		It("should send large message over UDP", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, join(setSyslogSpecValues, func(output *obs.OutputSpec) {
					output.Syslog.URL = "udp://127.0.0.1:24224"
				}))
			Expect(framework.Deploy()).To(BeNil())

			var MaxLen int = 30000
			Expect(framework.WritesNApplicationLogsOfSize(1, MaxLen, 0)).To(BeNil())
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " - ")
			payload := strings.TrimSpace(fields[1])
			record := map[string]interface{}{}
			Expect(json.Unmarshal([]byte(payload), &record)).To(BeNil(), fmt.Sprintf("payload: %q", payload))
			msg := record["message"]
			var message string
			message, ok := msg.(string)
			Expect(ok).To(BeTrue())
			ReceivedLen := uint64(len(message))
			Expect(ReceivedLen).To(BeEquivalentTo(MaxLen), "Expected the message length to be the same")
		})
		It("should send NonJson App logs to syslog", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, setSyslogSpecValues)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range NonJsonAppLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})
		It("should take values of appname, procid, messageid from record", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, join(setSyslogSpecValues, func(spec *obs.OutputSpec) {
					spec.Syslog.AppName = `{.appname_key||"none"}`
					spec.Syslog.ProcId = `{.procid_key||"none"}`
					spec.Syslog.MsgId = `{.msgid_key||"none"}`
				}))
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("rec_appname"))
			Expect(getProcID(fields)).To(Equal("rec_procid"))
			Expect(getMsgID(fields)).To(Equal("rec_msgid"))
		})
		It("should allow combination of static + dynamic setting of appname, procid, messageid from record", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, join(setSyslogSpecValues, func(spec *obs.OutputSpec) {
					spec.Syslog.AppName = `foo-{.openshift.cluster_id||"none"}`
					spec.Syslog.ProcId = `bar-{.appname_key||"none"}`
					spec.Syslog.MsgId = `baz{.level||"none"}.{.log_type||"none"}`
				}))
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("foo-functional"))
			Expect(getProcID(fields)).To(Equal("bar-rec_appname"))
			Expect(getMsgID(fields)).To(Equal("bazdefault.application"))
		})
		It("should send logs with delivery mode `atLeastOnce` configured", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, func(output *obs.OutputSpec) {
					output.Syslog.Tuning = &obs.SyslogTuningSpec{
						DeliveryMode: obs.DeliveryModeAtLeastOnce,
					}
				})
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty(), "Expected the receiver to receive the message")
		})
	})
	Context("Audit logs", func() {
		It("should send kubernetes audit logs", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToSyslogOutput(obs.SyslogRFC5424, setSyslogSpecValues)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			Expect(framework.WriteK8sAuditLog(1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeSyslog))
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
		It("should send openshift audit logs", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToSyslogOutput(obs.SyslogRFC5424, setSyslogSpecValues)
			Expect(framework.Deploy()).To(Succeed())

			// Log message data
			Expect(framework.WriteOpenshiftAuditLog(1)).To(Succeed())

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
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

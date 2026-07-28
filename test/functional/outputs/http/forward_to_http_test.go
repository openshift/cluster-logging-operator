package http

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	"k8s.io/apimachinery/pkg/api/resource"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

var _ = Describe("[Functional][Outputs][Http] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput(func(output *obs.OutputSpec) {
				output.HTTP.Tuning = &obs.HTTPTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						DeliveryMode:     obs.DeliveryModeAtLeastOnce,
						MaxRetryDuration: utils.GetPtr(time.Duration(30)),
						MinRetryDuration: utils.GetPtr(time.Duration(5)),
						MaxWrite:         utils.GetPtr(resource.MustParse("1M")),
					},
				}
			})
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
			userName := "imauser"
			password := "iwonttell"
			secretName := "mysecrets"
			framework.Forwarder.Spec.Outputs[0].HTTP.Authentication = &obs.HTTPAuthentication{
				Username: &obs.SecretReference{
					Key:        "username",
					SecretName: secretName,
				},
				Password: &obs.SecretReference{
					Key:        "password",
					SecretName: secretName,
				},
			}
			framework.Secrets = append(framework.Secrets, runtime.NewSecret(framework.Namespace, secretName, map[string][]byte{
				"username": []byte(userName),
				"password": []byte(password),
			}))

			return func(b *runtime.PodBuilder) error {
				return f.AddVectorHttpOutput(b, f.Forwarder.Spec.Outputs[0], functional.Option{Name: "username", Value: userName}, functional.Option{Name: "password", Value: password})
			}
		}),
		Entry("should send message over http to fluentd", func(f *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddFluentdHttpOutput(b, f.Forwarder.Spec.Outputs[0])
			}
		}),
	)

	Context("with journal logs", func() {
		It("should populate hostname for journal logs", func() {
			framework = functional.NewCollectorFunctionalFramework()
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeInfrastructure).
				ToHttpOutput()

			Expect(framework.Deploy()).To(BeNil())

			logline := functional.NewJournalLog(6, "test journal message", "functional-test-node")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			raw, err := framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).ToNot(BeEmpty())

			var logs []types.JournalLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(logs).ToNot(BeEmpty())
			Expect(logs[0].Hostname).ToNot(BeEmpty(), "Expected hostname to be populated for journal logs")
		})
	})

	Context("with tuning parameters", func() {
		DescribeTable("with compression", func(compression string) {
			framework.Forwarder.Spec.Outputs[0].HTTP.Tuning = &obs.HTTPTuningSpec{
				Compression: compression,
			}

			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				func(b *runtime.PodBuilder) error {
					return framework.AddVectorHttpOutput(b, framework.Forwarder.Spec.Outputs[0])
				},
				func(builder *runtime.PodBuilder) error {
					builder.AddLabels(map[string]string{
						"app.kubernetes.io/name": "somevalue",
						"foo.bar":                "a123",
					})
					return nil
				},
			})).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())
		},
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with snappy", "snappy"),
			Entry("should pass with zlib", "zlib"),
			Entry("should pass with no compression", "none"))
	})

	// LOG-9386: Verify that deeply nested JSON events (>32 protobuf nesting levels)
	// don't crash Vector when disk buffering is enabled (AtLeastOnce delivery mode).
	// Requires JSON parse filter so the nested JSON becomes a nested object in Vector.
	Context("with deeply nested JSON and disk buffering", func() {
		It("should drop over-nested events gracefully without crashing (LOG-9386)", func() {
			f := functional.NewCollectorFunctionalFramework()
			defer f.Cleanup()
			obstestruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithParseJson().
				ToHttpOutput(func(output *obs.OutputSpec) {
					output.HTTP.Tuning = &obs.HTTPTuningSpec{
						BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
							DeliveryMode: obs.DeliveryModeAtLeastOnce,
						},
					}
				})

			Expect(f.DeployWithVisitors([]runtime.PodBuilderVisitor{
				func(b *runtime.PodBuilder) error {
					return f.AddVectorHttpOutput(b, f.Forwarder.Spec.Outputs[0])
				},
			})).To(BeNil())

			// Generate 40-level deep nested JSON (protobuf decode limit is 32)
			nested := `{"msg":"deep"}`
			for i := 0; i < 40; i++ {
				nested = fmt.Sprintf(`{"l%d":%s}`, i, nested)
			}
			deepMsg := functional.CreateAppLogFromJson(nested)
			Expect(f.WriteMessagesToApplicationLog(deepMsg, 1)).To(BeNil())

			// Write a normal message to confirm the pipeline still works after dropping the nested one
			marker := fmt.Sprintf("normal-log-%d", time.Now().UnixNano())
			normalMsg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), fmt.Sprintf(`{"msg":"%s"}`, marker), false)
			Expect(f.WriteMessagesToApplicationLog(normalMsg, 1)).To(BeNil())

			raw, err := f.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			output := strings.Join(raw, "\n")
			Expect(output).To(ContainSubstring(marker),
				"Expected the normal message to be delivered")
			Expect(output).ToNot(ContainSubstring("l39"),
				"Expected the deeply nested event to be dropped, not delivered")

			collectorLogs, err := f.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLogs).ToNot(ContainSubstring("InvalidProtobufPayload"),
				"Vector should not crash with InvalidProtobufPayload on deeply nested events")
			Expect(collectorLogs).ToNot(ContainSubstring("failed to decoded record"),
				"Vector should not fail to decode buffered records")
			Expect(collectorLogs).To(ContainSubstring("Events dropped"),
				"Expected over-nested event to be reported as dropped")
			Expect(collectorLogs).To(ContainSubstring("Event nesting cost exceeds maximum"),
				"Expected nesting cost warning for the over-nested event")
		})
	})

	Context("timestamp in audit logs", func() {
		DescribeTable("audit log should have a valid timestamp",
			func(writeLog func() error, logSource obs.AuditSource) {
				framework = functional.NewCollectorFunctionalFramework()
				obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeAudit).
					ToHttpOutput()

				Expect(framework.Deploy()).To(Succeed())

				Expect(writeLog()).To(Succeed())

				logs, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeHTTP))
				Expect(err).To(Succeed())
				Expect(logs).ToNot(BeEmpty())
				var auditLogs []types.AuditLogCommon
				jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
				Expect(types.ParseLogsFrom(jsonString, &auditLogs, false)).To(Succeed())
				Expect(auditLogs).To(HaveLen(1))
				Expect(auditLogs[0].LogType).To(Equal(string(obs.InputTypeAudit)))
				Expect(auditLogs[0].LogSource).To(Equal(string(logSource)))
				Expect(auditLogs[0].Timestamp).ToNot(BeZero())
			},
			Entry("kubernetes audit log", func() error {
				return framework.WriteK8sAuditLog(1)
			}, obs.AuditSourceKube),
			Entry("OpenShift audit log", func() error {
				return framework.WriteOpenshiftAuditLog(1)
			}, obs.AuditSourceOpenShift),
		)
	})

	// Verify that component_sent_bytes_total carries component_id for the HTTP
	// output. This serves as a positive control alongside the CloudWatch
	// regression test for LOG-7893 — the HTTP sink emits bytes metrics from
	// within the Driver's future context (not a spawned buffer worker), so its
	// labels should always be correct.
	Context("When checking collector metrics for HTTP output", func() {
		BeforeEach(func() {
			role, binding, tokenBinding, err := framework.SetupMetricsRBAC()
			Expect(err).To(Succeed())
			DeferCleanup(func() {
				_ = framework.Test.Delete(tokenBinding)
				_ = framework.Test.Delete(binding)
				_ = framework.Test.Delete(role)
			})
		})

		It("should emit component_sent_bytes_total with component_id label", func() {
			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				func(b *runtime.PodBuilder) error {
					return framework.AddVectorHttpOutput(b, framework.Forwarder.Spec.Outputs[0])
				},
			})).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "metrics test message", false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 10)).To(BeNil())

			lines, err := framework.CollectMetricLines("component_sent_bytes_total", `component_id="output_http"`, 30*time.Second)
			Expect(err).To(BeNil(), "Timed out waiting for component_sent_bytes_total with component_id label")

			for _, line := range lines {
				log.V(2).Info("component_sent_bytes_total line", "line", line)
				Expect(line).To(ContainSubstring(`component_id=`),
					"component_sent_bytes_total without component_id label (transport-layer duplicate): %s", line)
			}
		})
	})
})

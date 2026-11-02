package lokistack_test

//
//import (
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
//	"github.com/openshift/cluster-logging-operator/internal/api/observability"
//	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
//	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/output/lokistack"
//	"github.com/openshift/cluster-logging-operator/internal/utils"
//	corev1 "k8s.io/api/core/v1"
//)
//
//var _ = Describe("LokiStack Output", func() {
//
//	var op = utils.NoOptions
//
//	Context("New", func() {
//		It("should create exporters for application tenant", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//					Authentication: &obs.LokiStackAuthentication{
//						Token: &obs.BearerToken{
//							From: obs.BearerTokenFromSecret,
//						},
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//					Application: &obs.Application{
//						Includes: []obs.NamespaceContainerSpec{
//							{Namespace: "my-app"},
//						},
//					},
//				},
//			}
//
//			secrets := observability.Secrets{
//				"secret-name.token": &corev1.Secret{
//					Data: map[string][]byte{
//						"token": []byte("test-token"),
//					},
//				},
//			}
//
//			exporterIDs, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, secrets, op)
//
//			Expect(exporterIDs).To(HaveKey("application"))
//			Expect(exportersMap).To(HaveLen(1))
//
//			appExporterID := exporterIDs["application"]
//			Expect(appExporterID).To(Equal("output_test_output_application"))
//
//			exporter, ok := exportersMap[appExporterID].(*exporters.OtlpHttp)
//			Expect(ok).To(BeTrue())
//			Expect(exporter).ToNot(BeNil())
//			Expect(exporter.Endpoint).To(ContainSubstring("https://"))
//			Expect(exporter.Endpoint).To(ContainSubstring("logging-loki-gateway-http.openshift-logging.svc:8080"))
//			Expect(exporter.Endpoint).To(ContainSubstring("/api/logs/v1/application/otlp/v1/logs"))
//		})
//
//		It("should create exporters for all three tenants", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//				},
//				{
//					Name: "infra-logs",
//					Type: obs.InputTypeInfrastructure,
//				},
//				{
//					Name: "audit-logs",
//					Type: obs.InputTypeAudit,
//				},
//			}
//
//			exporterIDs, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			Expect(exporterIDs).To(HaveLen(3))
//			Expect(exporterIDs).To(HaveKey("application"))
//			Expect(exporterIDs).To(HaveKey("infrastructure"))
//			Expect(exporterIDs).To(HaveKey("audit"))
//
//			Expect(exportersMap).To(HaveLen(3))
//
//			// Verify application exporter
//			appExporter, ok := exportersMap[exporterIDs["application"]].(*exporters.OtlpHttp)
//			Expect(ok).To(BeTrue())
//			Expect(appExporter.Endpoint).To(ContainSubstring("/api/logs/v1/application/otlp/v1/logs"))
//
//			// Verify infrastructure exporter
//			infraExporter, ok := exportersMap[exporterIDs["infrastructure"]].(*exporters.OtlpHttp)
//			Expect(ok).To(BeTrue())
//			Expect(infraExporter.Endpoint).To(ContainSubstring("/api/logs/v1/infrastructure/otlp/v1/logs"))
//
//			// Verify audit exporter
//			auditExporter, ok := exportersMap[exporterIDs["audit"]].(*exporters.OtlpHttp)
//			Expect(ok).To(BeTrue())
//			Expect(auditExporter.Endpoint).To(ContainSubstring("/api/logs/v1/audit/otlp/v1/logs"))
//		})
//
//		It("should configure TLS when specified", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//				},
//				TLS: &obs.OutputTLSSpec{
//					InsecureSkipVerify: true,
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//				},
//			}
//
//			_, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			Expect(exportersMap).To(HaveLen(1))
//			for _, exp := range exportersMap {
//				exporter, ok := exp.(*exporters.OtlpHttp)
//				Expect(ok).To(BeTrue())
//				Expect(exporter.TLS).ToNot(BeNil())
//				Expect(exporter.TLS.InsecureSkipVerify).To(BeTrue())
//			}
//		})
//
//		It("should configure compression from tuning", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//					Tuning: &obs.LokiTuningSpec{
//						Compression: "gzip",
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//				},
//			}
//
//			_, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			for _, exp := range exportersMap {
//				exporter, ok := exp.(*exporters.OtlpHttp)
//				Expect(ok).To(BeTrue())
//				Expect(exporter.Compression).To(Equal("gzip"))
//			}
//		})
//
//		It("should configure retry and queue for at-least-once delivery", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//					Tuning: &obs.LokiTuningSpec{
//						BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
//							DeliveryMode: obs.DeliveryModeAtLeastOnce,
//						},
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//				},
//			}
//
//			_, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			for _, exp := range exportersMap {
//				exporter, ok := exp.(*exporters.OtlpHttp)
//				Expect(ok).To(BeTrue())
//				Expect(exporter.RetryOnFailure).ToNot(BeNil())
//				Expect(exporter.RetryOnFailure.Enabled).To(BeTrue())
//				Expect(exporter.SendingQueue).ToNot(BeNil())
//				Expect(exporter.SendingQueue.Enabled).To(BeTrue())
//			}
//		})
//
//		It("should handle HTTP receiver as audit tenant", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "http-receiver",
//					Type: obs.InputTypeReceiver,
//					Receiver: &obs.ReceiverSpec{
//						Type: obs.ReceiverTypeHTTP,
//						Port: 8080,
//						HTTP: &obs.HTTPReceiver{
//							Format: obs.HTTPReceiverFormatKubeAPIAudit,
//						},
//					},
//				},
//			}
//
//			exporterIDs, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			Expect(exporterIDs).To(HaveKey("audit"))
//			Expect(exportersMap).To(HaveLen(1))
//
//			auditExporter, ok := exportersMap[exporterIDs["audit"]].(*exporters.OtlpHttp)
//			Expect(ok).To(BeTrue())
//			Expect(auditExporter.Endpoint).To(ContainSubstring("/api/logs/v1/audit/otlp/v1/logs"))
//		})
//
//		It("should handle syslog receiver as infrastructure tenant", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "syslog-receiver",
//					Type: obs.InputTypeReceiver,
//					Receiver: &obs.ReceiverSpec{
//						Type: obs.ReceiverTypeSyslog,
//						Port: 514,
//					},
//				},
//			}
//
//			exporterIDs, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			Expect(exporterIDs).To(HaveKey("infrastructure"))
//			Expect(exportersMap).To(HaveLen(1))
//
//			infraExporter, ok := exportersMap[exporterIDs["infrastructure"]].(*exporters.OtlpHttp)
//			Expect(ok).To(BeTrue())
//			Expect(infraExporter.Endpoint).To(ContainSubstring("/api/logs/v1/infrastructure/otlp/v1/logs"))
//		})
//
//		It("should map snappy compression to gzip (OTLP limitation)", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//					Tuning: &obs.LokiTuningSpec{
//						Compression: "snappy",
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//				},
//			}
//
//			_, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, nil)
//
//			for _, exp := range exportersMap {
//				exporter, ok := exp.(*exporters.OtlpHttp)
//				Expect(ok).To(BeTrue())
//				// Snappy is not supported by OTLP, should fallback to gzip
//				Expect(exporter.Compression).To(Equal("gzip"))
//			}
//		})
//	})
//
//	Context("determineTenants", func() {
//		It("should include infrastructure tenant when application includes infra namespaces", func() {
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-with-infra",
//					Type: obs.InputTypeApplication,
//					Application: &obs.Application{
//						Includes: []obs.NamespaceContainerSpec{
//							{Namespace: "my-app"},
//							{Namespace: "openshift-*"}, // Infra namespace pattern
//						},
//					},
//				},
//			}
//
//			outputSpec := obs.OutputSpec{
//				Name: "test",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "loki",
//						Namespace: "ns",
//					},
//				},
//			}
//
//			exporterIDs, _ := lokistack.New("test", outputSpec, inputSpecs, nil)
//
//			// Should have both application and infrastructure
//			Expect(exporterIDs).To(HaveKey("application"))
//			Expect(exporterIDs).To(HaveKey("infrastructure"))
//		})
//	})
//
//	Context("Bearer Token Authentication", func() {
//		It("should configure bearer token from secret", func() {
//			outputSpec := obs.OutputSpec{
//				Name: "lokistack-output",
//				Type: obs.OutputTypeLokiStack,
//				LokiStack: &obs.LokiStack{
//					Target: obs.LokiStackTarget{
//						Name:      "logging-loki",
//						Namespace: "openshift-logging",
//					},
//					Authentication: &obs.LokiStackAuthentication{
//						Token: &obs.BearerToken{
//							From: obs.BearerTokenFromSecret,
//							Secret: &obs.BearerTokenSecretKey{
//								Name: "token-secret",
//								Key:  "token",
//							},
//						},
//					},
//				},
//			}
//
//			inputSpecs := []obs.InputSpec{
//				{
//					Name: "app-logs",
//					Type: obs.InputTypeApplication,
//				},
//			}
//
//			secrets := observability.Secrets{
//				"token-secret.token": &corev1.Secret{
//					Data: map[string][]byte{
//						"token": []byte("my-bearer-token"),
//					},
//				},
//			}
//
//			_, exportersMap := lokistack.New("test-output", outputSpec, inputSpecs, secrets)
//
//			for _, exp := range exportersMap {
//				exporter, ok := exp.(*exporters.OtlpHttp)
//				Expect(ok).To(BeTrue())
//				Expect(exporter.Headers).ToNot(BeNil())
//				Expect(exporter.Headers).To(HaveKey("Authorization"))
//				Expect(exporter.Headers["Authorization"]).To(Equal("Bearer my-bearer-token"))
//			}
//		})
//	})
//})

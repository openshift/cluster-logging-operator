package network

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Network Ports", func() {
	DescribeTable("parsePortProtocolFromURL",
		func(url string, expected *factory.PortProtocol) {
			port := parsePortProtocolFromURL(url)
			Expect(port).To(Equal(expected))
		},
		// Valid URLs with ports
		Entry("should extract the port from HTTP URL", "http://example.com:8080", &factory.PortProtocol{Port: 8080, Protocol: corev1.ProtocolTCP}),
		Entry("should extract the port from HTTPS URL", "https://example.com:9200", &factory.PortProtocol{Port: 9200, Protocol: corev1.ProtocolTCP}),
		Entry("should extract the port from TCP URL", "tcp://kafka.example.com:9092", &factory.PortProtocol{Port: 9092, Protocol: corev1.ProtocolTCP}),
		Entry("should extract the port from TLS URL", "tls://kafka.example.com:9093", &factory.PortProtocol{Port: 9093, Protocol: corev1.ProtocolTCP}),
		Entry("should extract the port from UDP URL", "udp://example.com:5140", &factory.PortProtocol{Port: 5140, Protocol: corev1.ProtocolUDP}),

		// URLs without explicit ports
		Entry("should return nil for HTTP URL without port", "http://example.com", (*factory.PortProtocol)(nil)),
		Entry("should return nil for HTTPS URL without port", "https://example.com", (*factory.PortProtocol)(nil)),

		// Invalid URLs
		Entry("should return nil for malformed URL", "not-a-url", (*factory.PortProtocol)(nil)),
		Entry("should return nil for empty string", "", (*factory.PortProtocol)(nil)),
		Entry("should return nil for invalid port", "http://example.com:invalid", (*factory.PortProtocol)(nil)),
		Entry("should return nil for negative port", "http://example.com:-80", (*factory.PortProtocol)(nil)),
	)

	DescribeTable("getDefaultPort",
		func(outputType obs.OutputType, urlStr string, expected int32) {
			port := getDefaultOutputPort(outputType, urlStr)
			Expect(port).To(Equal(expected))
		},
		// Different output types
		Entry("should return 9200 for Elasticsearch", obs.OutputTypeElasticsearch, "", int32(9200)),
		Entry("should return 8088 for Splunk", obs.OutputTypeSplunk, "", int32(8088)),
		Entry("should return 3100 for Loki", obs.OutputTypeLoki, "", int32(3100)),
		Entry("should return 514 for Syslog", obs.OutputTypeSyslog, "", int32(514)),
		Entry("should return 4318 for OTLP", obs.OutputTypeOTLP, "", int32(4318)),
		Entry("should return 443 for Cloudwatch", obs.OutputTypeCloudwatch, "", int32(443)),
		Entry("should return 443 for AzureMonitor", obs.OutputTypeAzureMonitor, "", int32(443)),
		Entry("should return 443 for GoogleCloudLogging", obs.OutputTypeGoogleCloudLogging, "", int32(443)),
		Entry("should return 8080 for LokiStack", obs.OutputTypeLokiStack, "", int32(8080)),
		Entry("should return 443 for S3", obs.OutputTypeS3, "", int32(443)),

		// Kafka with different schemes
		Entry("should return 9092 for plaintext Kafka", obs.OutputTypeKafka, "tcp://kafka.example.com", int32(9092)),
		Entry("should return 9093 for TLS Kafka", obs.OutputTypeKafka, "tls://kafka.example.com", int32(9093)),
		Entry("should return 9092 for Kafka with empty URL", obs.OutputTypeKafka, "", int32(9092)),

		// HTTP with different schemes
		Entry("should return 80 for HTTP scheme", obs.OutputTypeHTTP, "http://example.com", int32(80)),
		Entry("should return 443 for HTTPS scheme", obs.OutputTypeHTTP, "https://example.com", int32(443)),
		Entry("should return 443 for HTTP with no scheme", obs.OutputTypeHTTP, "", int32(443)),
	)

	It("should not panic for all supported output types", func() {
		for _, outputType := range obs.OutputTypes {
			Expect(func() { getDefaultOutputPort(outputType, "") }).ToNot(Panic())
		}
	})

	It("should panic for unknown output type", func() {
		Expect(func() { getDefaultOutputPort(obs.OutputType("unknown"), "") }).To(Panic())
	})

	DescribeTable("getKafkaBrokerPortProtocols",
		func(brokers []obs.BrokerURL, expectedPorts []factory.PortProtocol) {
			ports := getKafkaBrokerPortProtocols(brokers)
			Expect(ports).To(ConsistOf(expectedPorts))
		},
		Entry("should extract ports from multiple brokers",
			[]obs.BrokerURL{
				"tcp://broker1:9092",
				"tcp://broker2:9093",
				"tls://broker3:9094",
			},
			[]factory.PortProtocol{
				{Port: 9092, Protocol: corev1.ProtocolTCP},
				{Port: 9093, Protocol: corev1.ProtocolTCP},
				{Port: 9094, Protocol: corev1.ProtocolTCP},
			},
		),
		Entry("should handle single broker",
			[]obs.BrokerURL{"tcp://broker1:9092"},
			[]factory.PortProtocol{{Port: 9092, Protocol: corev1.ProtocolTCP}},
		),
		Entry("should handle single broker with no port",
			[]obs.BrokerURL{"tcp://broker1"},
			[]factory.PortProtocol{{Port: 9092, Protocol: corev1.ProtocolTCP}},
		),
		Entry("should handle mixed brokers with and without ports",
			[]obs.BrokerURL{
				"tcp://broker1:9094",
				"tcp://broker2",
				"tls://broker3",
			},
			[]factory.PortProtocol{
				{Port: 9094, Protocol: corev1.ProtocolTCP},
				{Port: 9092, Protocol: corev1.ProtocolTCP},
				{Port: 9093, Protocol: corev1.ProtocolTCP},
			},
		),
		Entry("should return empty ports for empty brokers",
			[]obs.BrokerURL{},
			[]factory.PortProtocol{},
		),
	)

	Context("getPortProtocolFromOutputURL", func() {
		DescribeTable("Elasticsearch",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type:          obs.OutputTypeElasticsearch,
					Elasticsearch: &obs.Elasticsearch{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from Elasticsearch URL",
				"https://es.example.com:9500",
				[]factory.PortProtocol{{Port: 9500, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port when no port in URL",
				"https://es.example.com",
				[]factory.PortProtocol{{Port: 9200, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("Splunk",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type:   obs.OutputTypeSplunk,
					Splunk: &obs.Splunk{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from Splunk URL",
				"https://splunk.example.com:8000",
				[]factory.PortProtocol{{Port: 8000, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for Splunk without explicit port",
				"https://splunk.example.com",
				[]factory.PortProtocol{{Port: 8088, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("Loki",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeLoki,
					Loki: &obs.Loki{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from Loki URL",
				"https://loki.example.com:3500",
				[]factory.PortProtocol{{Port: 3500, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for Loki without explicit port",
				"https://loki.example.com",
				[]factory.PortProtocol{{Port: 3100, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("Syslog",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type:   obs.OutputTypeSyslog,
					Syslog: &obs.Syslog{URL: urlStr},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from Syslog URL",
				"tcp://syslog.example.com:500",
				[]factory.PortProtocol{{Port: 500, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for Syslog without explicit port",
				"tcp://syslog.example.com",
				[]factory.PortProtocol{{Port: 514, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("OTLP",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeOTLP,
					OTLP: &obs.OTLP{URL: urlStr},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from OTLP URL",
				"http://otlp.example.com:4500",
				[]factory.PortProtocol{{Port: 4500, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for OTLP without explicit port",
				"http://otlp.example.com",
				[]factory.PortProtocol{{Port: 4318, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("HTTP",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeHTTP,
					HTTP: &obs.HTTP{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from HTTP URL",
				"http://http.example.com:8000",
				[]factory.PortProtocol{{Port: 8000, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for HTTP without explicit port",
				"http://http.example.com",
				[]factory.PortProtocol{{Port: 80, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for HTTPS without explicit port",
				"https://http.example.com",
				[]factory.PortProtocol{{Port: 443, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("Cloudwatch",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type:       obs.OutputTypeCloudwatch,
					Cloudwatch: &obs.Cloudwatch{URL: urlStr},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from Cloudwatch URL",
				"https://cloudwatch.amazonaws.com:8000",
				[]factory.PortProtocol{{Port: 8000, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for Cloudwatch without explicit port",
				"https://cloudwatch.amazonaws.com",
				[]factory.PortProtocol{{Port: 443, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("S3",
			func(urlStr string, expected []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeS3,
					S3:   &obs.S3{URL: urlStr},
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(Equal(expected))
			},
			Entry("should extract port from S3 URL",
				"https://some-s3-bucket.com:5555",
				[]factory.PortProtocol{{Port: 5555, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should use default port for S3 without explicit port",
				"https://s3.amazonaws.com",
				[]factory.PortProtocol{{Port: 443, Protocol: corev1.ProtocolTCP}},
			),
		)

		DescribeTable("Kafka",
			func(kafka *obs.Kafka, expectedPorts []factory.PortProtocol) {
				output := obs.OutputSpec{
					Type:  obs.OutputTypeKafka,
					Kafka: kafka,
				}
				ports := getPortProtocolFromOutputURL(output)
				Expect(ports).To(ConsistOf(expectedPorts))
			},
			Entry("should extract port from Kafka URL",
				&obs.Kafka{URL: "tcp://kafka.example.com:9000"},
				[]factory.PortProtocol{{Port: 9000, Protocol: corev1.ProtocolTCP}},
			),
			Entry("should extract ports from Kafka brokers when URL is empty",
				&obs.Kafka{
					URL: "",
					Brokers: []obs.BrokerURL{
						"tcp://broker1:9100",
						"tls://broker2:9200",
					},
				},
				[]factory.PortProtocol{
					{Port: 9100, Protocol: corev1.ProtocolTCP},
					{Port: 9200, Protocol: corev1.ProtocolTCP},
				},
			),
		)
	})

	Describe("GetOutputPortsWithProtocols", func() {
		Context("when given multiple outputs", func() {
			It("should extract unique ports from multiple outputs", func() {
				outputs := []obs.OutputSpec{
					{
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{URL: "https://es.example.com:9200"},
						},
					},
					{
						Type: obs.OutputTypeSplunk,
						Splunk: &obs.Splunk{
							URLSpec: obs.URLSpec{URL: "https://splunk.example.com:8088"},
						},
					},
					{
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{URL: "http://loki.example.com:3100"},
						},
					},
				}
				portMap := map[factory.PortProtocol]bool{}
				GetOutputPortsWithProtocols(outputs, portMap)
				Expect(portMap).To(Equal(map[factory.PortProtocol]bool{
					{Port: 9200, Protocol: corev1.ProtocolTCP}: true,
					{Port: 8088, Protocol: corev1.ProtocolTCP}: true,
					{Port: 3100, Protocol: corev1.ProtocolTCP}: true,
				}))
			})

			It("should deduplicate ports from multiple outputs", func() {
				outputs := []obs.OutputSpec{
					{
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{URL: "https://es1.example.com:9200"},
						},
					},
					{
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{URL: "https://es2.example.com:9200"},
						},
					},
					{
						Type: obs.OutputTypeSplunk,
						Splunk: &obs.Splunk{
							URLSpec: obs.URLSpec{URL: "https://splunk.example.com:8088"},
						},
					},
				}
				portMap := map[factory.PortProtocol]bool{}
				GetOutputPortsWithProtocols(outputs, portMap)
				Expect(portMap).To(Equal(map[factory.PortProtocol]bool{
					{Port: 9200, Protocol: corev1.ProtocolTCP}: true,
					{Port: 8088, Protocol: corev1.ProtocolTCP}: true,
				}))
			})

			It("should handle outputs with default ports", func() {
				outputs := []obs.OutputSpec{
					{
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{URL: "https://es.example.com"}, // no port, should use default 9200
						},
					},
					{
						Type: obs.OutputTypeHTTP,
						HTTP: &obs.HTTP{
							URLSpec: obs.URLSpec{URL: "http://api.example.com"}, // no port, should use default 80
						},
					},
				}
				portMap := map[factory.PortProtocol]bool{}
				GetOutputPortsWithProtocols(outputs, portMap)
				Expect(portMap).To(Equal(map[factory.PortProtocol]bool{
					{Port: 9200, Protocol: corev1.ProtocolTCP}: true,
					{Port: 80, Protocol: corev1.ProtocolTCP}:   true,
				}))
			})

			It("should handle complex Kafka output with multiple brokers", func() {
				outputs := []obs.OutputSpec{
					{
						Type: obs.OutputTypeKafka,
						Kafka: &obs.Kafka{
							URL: "",
							Brokers: []obs.BrokerURL{
								"tcp://broker1:9092",
								"tls://broker2:9093",
								"tcp://broker3:9092", // duplicate port
							},
						},
					},
					{
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{URL: "https://es.example.com:9200"},
						},
					},
				}
				portMap := map[factory.PortProtocol]bool{}
				GetOutputPortsWithProtocols(outputs, portMap)
				Expect(portMap).To(Equal(map[factory.PortProtocol]bool{
					{Port: 9092, Protocol: corev1.ProtocolTCP}: true,
					{Port: 9093, Protocol: corev1.ProtocolTCP}: true,
					{Port: 9200, Protocol: corev1.ProtocolTCP}: true,
				}))
			})
		})
	})

	Describe("GetInputPorts", func() {
		Context("with input receiver specs", func() {
			It("should extract ports from HTTP receivers", func() {
				inputs := []obs.InputSpec{
					{
						Name: "http-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
						},
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(ConsistOf(int32(8080)))
			})

			It("should extract ports from syslog receivers", func() {
				inputs := []obs.InputSpec{
					{
						Name: "syslog-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeSyslog,
							Port: 5140,
						},
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(ConsistOf(int32(5140)))
			})

			It("should extract unique ports from multiple receivers", func() {
				inputs := []obs.InputSpec{
					{
						Name: "http-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
						},
					},
					{
						Name: "syslog-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeSyslog,
							Port: 5140,
						},
					},
					{
						Name: "another-http-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 5000, // duplicate port
						},
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(ConsistOf(int32(8080), int32(5140), int32(5000)))
			})

			It("should ignore non-receiver input types", func() {
				inputs := []obs.InputSpec{
					{
						Name: "app-input",
						Type: obs.InputTypeApplication,
					},
					{
						Name: "infra-input",
						Type: obs.InputTypeInfrastructure,
					},
					{
						Name: "audit-input",
						Type: obs.InputTypeAudit,
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(BeEmpty())
			})

			It("should ignore receiver inputs with nil receiver spec", func() {
				inputs := []obs.InputSpec{
					{
						Name:     "receiver-input",
						Type:     obs.InputTypeReceiver,
						Receiver: nil,
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(BeEmpty())
			})

			It("should ignore receiver inputs with zero port", func() {
				inputs := []obs.InputSpec{
					{
						Name: "receiver-input",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 0,
						},
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(BeEmpty())
			})

			It("should handle empty input list", func() {
				ports := GetInputPorts([]obs.InputSpec{})
				Expect(ports).To(BeEmpty())
			})

			It("should handle mixed input types with receivers", func() {
				inputs := []obs.InputSpec{
					{
						Name: "app-input",
						Type: obs.InputTypeApplication,
					},
					{
						Name: "http-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
						},
					},
					{
						Name: "audit-input",
						Type: obs.InputTypeAudit,
					},
					{
						Name: "syslog-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeSyslog,
							Port: 5140,
						},
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(ConsistOf(int32(8080), int32(5140)))
			})
		})
	})

	Describe("GetProxyPorts", func() {
		var originalEnvVars map[string]string

		BeforeEach(func() {
			// Save original environment variables
			originalEnvVars = make(map[string]string)
			for _, envVar := range constants.ProxyEnvVars {
				originalEnvVars[envVar] = os.Getenv(envVar)
				os.Unsetenv(envVar) // Clear all proxy env vars for clean test state
			}
		})

		AfterEach(func() {
			// Restore original environment variables
			for envVar, value := range originalEnvVars {
				if value != "" {
					os.Setenv(envVar, value)
				} else {
					os.Unsetenv(envVar)
				}
			}
		})

		setProxies := func(httpProxy, httpsProxy string) map[string]string {
			return map[string]string{
				"http_proxy":  httpProxy,
				"https_proxy": httpsProxy,
			}
		}

		expectedProxyPorts := func(ports ...int32) []factory.PortProtocol {
			expectedPortProtocols := make([]factory.PortProtocol, 0, len(ports))
			for _, port := range ports {
				expectedPortProtocols = append(expectedPortProtocols, factory.PortProtocol{Port: port, Protocol: corev1.ProtocolTCP})
			}
			return expectedPortProtocols
		}

		DescribeTable("when proxy environment variables are set",
			func(envVars map[string]string, expectedPorts []factory.PortProtocol) {
				for key, value := range envVars {
					os.Setenv(key, value)
				}

				portMap := map[factory.PortProtocol]bool{}
				GetProxyPorts(portMap)
				ports := make([]factory.PortProtocol, 0, len(portMap))
				for pp := range portMap {
					ports = append(ports, pp)
				}
				if len(expectedPorts) == 0 {
					Expect(ports).To(BeEmpty())
				} else {
					Expect(ports).To(ConsistOf(expectedPorts))
				}
			},
			Entry("should extract ports from HTTP proxy URLs with explicit ports",
				setProxies("http://proxy.example.com:8080", "https://proxy.example.com:8443"),
				expectedProxyPorts(8080, 8443),
			),
			Entry("should use default ports when proxy URLs don't specify ports",
				setProxies("http://proxy.example.com", "https://proxy.example.com"),
				expectedProxyPorts(80, 443),
			),
			Entry("should return empty when proxy URLs have unknown schemes without ports",
				setProxies("proxy://proxy.example.com", "invalid://proxy.example.com"),
				expectedProxyPorts(),
			),
			Entry("should deduplicate identical proxy ports",
				setProxies("http://proxy.example.com:8080", "https://proxy.example.com:8080"),
				expectedProxyPorts(8080),
			),
			Entry("should handle malformed proxy URLs gracefully",
				setProxies("not-a-valid-url", "http://proxy.example.com:8080"),
				expectedProxyPorts(8080),
			),
			Entry("should extract explicitly specified standard ports from proxy URLs",
				setProxies("http://proxy.example.com:80", "https://proxy.example.com:443"),
				expectedProxyPorts(80, 443),
			),
			Entry("should handle proxy URLs with authentication",
				setProxies("http://user:password@proxy.example.com:8080", "https://user:password@proxy.example.com:8443"),
				expectedProxyPorts(8080, 8443),
			),
			Entry("should handle IPv6 proxy URLs",
				setProxies("http://[::1]:8080", ""),
				expectedProxyPorts(8080),
			),
			Entry("should handle empty string proxy URLs gracefully",
				setProxies("", "https://proxy.example.com:8443"),
				expectedProxyPorts(8443),
			),
			Entry("should handle invalid ports on urls gracefully",
				setProxies("http://proxy.example.com:invalid", "https://proxy.example.com:no-a-port"),
				expectedProxyPorts(),
			),
			Entry("should handle mix of uppercase and lowercase proxy environment variables",
				map[string]string{
					"HTTP_PROXY":  "http://proxy.example.com:8080",
					"https_proxy": "https://proxy.example.com:8443",
				},
				expectedProxyPorts(8080, 8443),
			),
		)

		It("should be empty portMap when no proxy environment variables are set", func() {
			portMap := map[factory.PortProtocol]bool{}
			GetProxyPorts(portMap)
			Expect(portMap).To(BeEmpty())
		})
	})
})

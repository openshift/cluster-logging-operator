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
	Context("parsePortProtocolFromURL", func() {
		DescribeTable("with valid URLs and standard URL schemes",
			func(url string, expectedPort int32) {
				var expected *factory.PortProtocol
				if expectedPort > 0 {
					expected = &factory.PortProtocol{Port: expectedPort, Protocol: corev1.ProtocolTCP}
				}
				Expect(parsePortProtocolFromURL(url)).To(Equal(expected))
			},
			// Valid URLs with ports
			Entry("should extract the port from HTTP URL", "http://example.com:8080", int32(8080)),
			Entry("should extract the port from HTTPS URL", "https://example.com:9200", int32(9200)),

			// Valid URLs without explicit ports
			Entry("should return 80 for HTTP URL without port", "http://example.com", constants.DefaultHTTPPort),
			Entry("should return 443 for HTTPS URL without port", "https://example.com", constants.DefaultHTTPSPort),

			// Special hosts w/ port
			Entry("should handle IPv4 address with port", "http://192.168.1.1:8080", int32(8080)),
			Entry("should handle IPv6 address with port", "http://[::1]:8080", int32(8080)),
			Entry("should handle IPv6 full address with port", "http://[2001:db8::1]:9200", int32(9200)),
			Entry("should handle localhost with port", "http://localhost:3000", int32(3000)),

			// Special hosts without port
			Entry("should handle IPv4 address without port", "http://192.168.1.1", constants.DefaultHTTPPort),
			Entry("should handle IPv6 address without port", "http://[::1]", constants.DefaultHTTPPort),
			Entry("should handle IPv6 full address without port", "http://[2001:db8::1]", constants.DefaultHTTPPort),
			Entry("should handle localhost without port", "http://localhost", constants.DefaultHTTPPort),

			Entry("should use well known scheme port for zero port", "http://example.com:0", constants.DefaultHTTPPort),

			// Should be nil for empty string
			Entry("should return nil for empty string", "", nil),
		)

		DescribeTable("with non-standard URL schemes with ports",
			func(url string, expectedPort int32, expectedProtocol corev1.Protocol) {
				Expect(parsePortProtocolFromURL(url)).To(Equal(&factory.PortProtocol{Port: expectedPort, Protocol: expectedProtocol}))
			},
			Entry("should extract the port from TCP URL with port", "tcp://kafka.example.com:9092", int32(9092), corev1.ProtocolTCP),
			Entry("should extract the port from TLS URL with port", "tls://kafka.example.com:9093", int32(9093), corev1.ProtocolTCP),
			Entry("should extract the port from UDP URL with port", "udp://example.com:5140", int32(5140), corev1.ProtocolUDP),
		)

		DescribeTable("with non-standard URL schemes without or invalid ports",
			func(url string) {
				Expect(func() { parsePortProtocolFromURL(url) }).To(Panic())
			},
			Entry("should panic for non standard URL scheme, tcp, without a port", "tcp://example.com"),
			Entry("should panic for non standard URL scheme, udp, without a port", "udp://example.com"),
			Entry("should panic for non standard URL scheme, tls, without a port", "tls://example.com"),
			Entry("should panic for non standard URL scheme, nonStandard, without a port", "nonStandard://example.com"),
			Entry("should panic for malformed URL", "not-a-url"),
			Entry("should panic for invalid port", "http://example.com:invalid"),
			Entry("should panic for negative port", "http://example.com:-80"),
		)
	})

	DescribeTable("getKafkaAndBrokerURLs",
		func(kafkaURL string, brokerURLs []obs.BrokerURL, expected []string) {
			kafka := obs.Kafka{
				Brokers: brokerURLs,
				URL:     kafkaURL,
			}
			Expect(getKafkaAndBrokerURLs(kafka)).To(Equal(expected))
		},
		Entry("should return URL when specified",
			"tcp://kafka.example.com:9092",
			nil,
			[]string{"tcp://kafka.example.com:9092"},
		),
		Entry("should return both URL and brokers when both are specified",
			"tcp://kafka.primary.com:9092",
			[]obs.BrokerURL{"tcp://broker1:9092", "tcp://broker2:9093"},
			[]string{"tcp://kafka.primary.com:9092", "tcp://broker1:9092", "tcp://broker2:9093"},
		),
		Entry("should return slice with empty string for nil brokers and empty URL",
			"",
			nil,
			[]string{""},
		),
		Entry("should return slice with empty string for empty brokers and empty URL",
			"",
			[]obs.BrokerURL{},
			[]string{""},
		),
		Entry("should return kafka.URL for empty brokers and non-empty URL",
			"tcp://kafka.example.com:9092",
			[]obs.BrokerURL{},
			[]string{"tcp://kafka.example.com:9092"},
		),
	)

	Context("getPortProtocolFromOutputURL", func() {
		makePortProtocol := func(port int32, protocol corev1.Protocol) []factory.PortProtocol {
			return []factory.PortProtocol{{Port: port, Protocol: protocol}}
		}

		// Helper to create TCP port protocols from port numbers
		makeTCPPorts := func(ports ...int32) []factory.PortProtocol {
			result := make([]factory.PortProtocol, len(ports))
			for i, port := range ports {
				result[i] = factory.PortProtocol{Port: port, Protocol: corev1.ProtocolTCP}
			}
			return result
		}

		DescribeTable("Elasticsearch",
			func(urlStr string, expectedPort int32) {
				output := obs.OutputSpec{
					Type:          obs.OutputTypeElasticsearch,
					Elasticsearch: &obs.Elasticsearch{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPort)))
			},
			Entry("should extract port from Elasticsearch URL",
				"https://es.example.com:9500", int32(9500)),
			Entry("should use default HTTPS scheme port when no port in URL",
				"https://es.example.com", constants.DefaultHTTPSPort),
			Entry("should use default HTTP scheme port when no port in URL",
				"http://es.example.com", constants.DefaultHTTPPort),
		)

		DescribeTable("Splunk",
			func(urlStr string, expectedPort int32) {
				output := obs.OutputSpec{
					Type:   obs.OutputTypeSplunk,
					Splunk: &obs.Splunk{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPort)))
			},
			Entry("should extract port from Splunk URL",
				"https://splunk.example.com:8000", int32(8000)),
			Entry("should use default HTTPS scheme port for Splunk without explicit port",
				"https://splunk.example.com", constants.DefaultHTTPSPort),
			Entry("should use default HTTP scheme port for Splunk without explicit port",
				"http://splunk.example.com", constants.DefaultHTTPPort),
		)

		DescribeTable("Loki",
			func(urlStr string, expectedPort int32) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeLoki,
					Loki: &obs.Loki{URLSpec: obs.URLSpec{URL: urlStr}},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPort)))
			},
			Entry("should extract port from Loki URL",
				"https://loki.example.com:3500", int32(3500)),
			Entry("should use default HTTPS scheme port for Loki without explicit port",
				"https://loki.example.com", constants.DefaultHTTPSPort),
			Entry("should use default HTTP scheme port for Loki without explicit port",
				"http://loki.example.com", constants.DefaultHTTPPort),
		)

		DescribeTable("Syslog",
			func(urlStr string, expectedPorts []factory.PortProtocol, shouldPanic bool) {
				output := obs.OutputSpec{
					Type:   obs.OutputTypeSyslog,
					Syslog: &obs.Syslog{URL: urlStr},
				}

				if shouldPanic {
					Expect(func() { getPortProtocolFromOutputURLs(output) }).To(Panic())
				} else {
					Expect(getPortProtocolFromOutputURLs(output)).To(Equal(expectedPorts))
				}
			},
			Entry("should extract port from Syslog TCP URL",
				"tcp://syslog.example.com:500", makeTCPPorts(500), false),
			Entry("should extract port from Syslog TLS URL",
				"tls://syslog.example.com:550", makeTCPPorts(550), false),
			Entry("should extract port from Syslog UDP URL",
				"udp://syslog.example.com:514", makePortProtocol(514, corev1.ProtocolUDP), false),

			// Non-standard URL schemes without ports should panic
			Entry("should panic for non standard URL scheme, tcp, without a port",
				"tcp://syslog.example.com", nil, true),
			Entry("should panic for non standard URL scheme, udp, without a port",
				"udp://syslog.example.com", nil, true),
			Entry("should panic for non standard URL scheme, tls, without a port",
				"tls://syslog.example.com", nil, true),
		)

		DescribeTable("OTLP",
			func(urlStr string, expectedPort int32) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeOTLP,
					OTLP: &obs.OTLP{URL: urlStr},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPort)))
			},
			Entry("should extract port from OTLP URL",
				"http://otlp.example.com:4500", int32(4500)),
			Entry("should use default HTTP scheme port for OTLP without explicit port",
				"http://otlp.example.com", constants.DefaultHTTPPort),
			Entry("should use default HTTPS scheme port for OTLP without explicit port",
				"https://otlp.example.com", constants.DefaultHTTPSPort),
		)

		DescribeTable("HTTP",
			func(urlStr string, proxyURLStr string, expectedPorts []int32) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeHTTP,
					HTTP: &obs.HTTP{
						URLSpec:  obs.URLSpec{URL: urlStr},
						ProxyURL: proxyURLStr,
					},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPorts...)))
			},
			Entry("should extract port from URL",
				"http://http.example.com:8000", "", []int32{8000}),
			Entry("should use default HTTP scheme port for HTTP without explicit port",
				"http://http.example.com", "", []int32{constants.DefaultHTTPPort}),
			Entry("should use default HTTPS scheme port for HTTPS without explicit port",
				"https://http.example.com", "", []int32{constants.DefaultHTTPSPort}),
			Entry("should extract ports from URL and ProxyURL when defined",
				"http://http.example.com:8000", "http://proxy.example.com:3128", []int32{8000, 3128}),
			Entry("should use default ports when URL and ProxyURL have no explicit ports",
				"http://http.example.com", "http://proxy.example.com", []int32{constants.DefaultHTTPPort, constants.DefaultHTTPPort}),
			Entry("should handle mixed schemes with URL and ProxyURL",
				"https://http.example.com:8443", "http://proxy.example.com:8080", []int32{8443, 8080}),
			Entry("should extract only URL port when ProxyURL is empty",
				"https://http.example.com:9443", "", []int32{9443}),
		)

		DescribeTable("Cloudwatch",
			func(urlStr string, expectedPort int32) {
				output := obs.OutputSpec{
					Type:       obs.OutputTypeCloudwatch,
					Cloudwatch: &obs.Cloudwatch{URL: urlStr},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPort)))
			},
			Entry("should extract port from Cloudwatch URL",
				"https://cloudwatch.amazonaws.com:8000", int32(8000)),
			Entry("should use default HTTP scheme port for Cloudwatch without explicit port",
				"http://cloudwatch.amazonaws.com", constants.DefaultHTTPPort),
			Entry("should use default HTTPS scheme port for Cloudwatch without explicit port",
				"https://cloudwatch.amazonaws.com", constants.DefaultHTTPSPort),
			Entry("should use default HTTPS port when URL is not defined",
				"", constants.DefaultHTTPSPort),
		)

		DescribeTable("S3",
			func(urlStr string, expectedPort int32) {
				output := obs.OutputSpec{
					Type: obs.OutputTypeS3,
					S3:   &obs.S3{URL: urlStr},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPort)))
			},
			Entry("should extract port from S3 URL",
				"https://some-s3-bucket.com:5555", int32(5555)),
			Entry("should use default HTTPS scheme port for S3 without explicit port",
				"https://s3.amazonaws.com", constants.DefaultHTTPSPort),
			Entry("should use default HTTP scheme port for S3 without explicit port",
				"http://s3.amazonaws.com", constants.DefaultHTTPPort),
			Entry("should use default HTTPS port when URL is not defined",
				"", constants.DefaultHTTPSPort),
		)

		DescribeTable("Kafka",
			func(urlStr string, brokers []obs.BrokerURL, expectedPorts []int32) {
				output := obs.OutputSpec{
					Type:  obs.OutputTypeKafka,
					Kafka: &obs.Kafka{URL: urlStr, Brokers: brokers},
				}
				Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(expectedPorts...)))
			},
			Entry("should extract port from Kafka URL",
				"tcp://kafka.example.com:9000",
				nil,
				[]int32{9000},
			),
			Entry("should extract ports from Kafka brokers when URL is empty",
				"",
				[]obs.BrokerURL{"tcp://broker1:9100", "tls://broker2:9200"},
				[]int32{9100, 9200},
			),
			Entry("should extract ports from both URL and brokers when both are provided",
				"tcp://kafka.example.com:9000",
				[]obs.BrokerURL{"tcp://broker1:9100", "tls://broker2:9200"},
				[]int32{9000, 9100, 9200},
			),
		)

		It("should return port 8080 for LokiStack", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeLokiStack,
			}
			Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(8080)))
		})

		It("should return default HTTPS port, 443, for Google Cloud Logging", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeGoogleCloudLogging,
			}
			Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(443)))
		})

		It("should return default HTTPS port, 443, for Azure Monitor", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeAzureMonitor,
			}
			Expect(getPortProtocolFromOutputURLs(output)).To(Equal(makeTCPPorts(443)))
		})

		It("should panic for unsupported output type", func() {
			output := obs.OutputSpec{
				Type: "unsupported",
			}
			Expect(func() { getPortProtocolFromOutputURLs(output) }).To(Panic())
		})
	})

	Describe("GetOutputPortsWithProtocols", func() {
		makeTCPPortsMap := func(ports ...int32) map[factory.PortProtocol]bool {
			result := make(map[factory.PortProtocol]bool)
			for _, port := range ports {
				result[factory.PortProtocol{Port: port, Protocol: corev1.ProtocolTCP}] = true
			}
			return result
		}

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
				expected := makeTCPPortsMap(9200, 8088, 3100)
				portMap := GetOutputPortsWithProtocols(outputs)
				Expect(portMap).To(Equal(expected))
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
				expected := makeTCPPortsMap(9200, 8088)
				Expect(GetOutputPortsWithProtocols(outputs)).To(Equal(expected))
			})

			It("should handle outputs with known schemes and no explicit ports", func() {
				outputs := []obs.OutputSpec{
					{
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{URL: "https://es.example.com"},
						},
					},
					{
						Type: obs.OutputTypeHTTP,
						HTTP: &obs.HTTP{
							URLSpec: obs.URLSpec{URL: "http://api.example.com"},
						},
					},
				}
				expected := makeTCPPortsMap(constants.DefaultHTTPSPort, constants.DefaultHTTPPort)
				Expect(GetOutputPortsWithProtocols(outputs)).To(Equal(expected))
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
				expected := makeTCPPortsMap(9092, 9093, 9200)
				Expect(GetOutputPortsWithProtocols(outputs)).To(Equal(expected))
			})
		})
	})

	Describe("GetInputPorts", func() {
		Context("with input receiver specs", func() {
			It("should extract ports from HTTP receivers", func() {
				inputs := []obs.InputSpec{
					{
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
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
						},
					},
					{
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeSyslog,
							Port: 5140,
						},
					},
					{
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 5000,
						},
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(ConsistOf(int32(8080), int32(5140), int32(5000)))
			})

			It("should ignore non-receiver input types", func() {
				inputs := []obs.InputSpec{
					{
						Type: obs.InputTypeApplication,
					},
					{
						Type: obs.InputTypeInfrastructure,
					},
					{
						Type: obs.InputTypeAudit,
					},
				}
				ports := GetInputPorts(inputs)
				Expect(ports).To(BeEmpty())
			})

			It("should ignore receiver inputs with nil receiver spec", func() {
				inputs := []obs.InputSpec{
					{
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
						Type: obs.InputTypeApplication,
					},
					{
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
						},
					},
					{
						Type: obs.InputTypeAudit,
					},
					{
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

				portMap := GetProxyPorts()
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
				expectedProxyPorts(8080, 8443)),
			Entry("should use default scheme ports when proxy URLs don't specify ports",
				setProxies("http://proxy.example.com", "https://proxy.example.com"),
				expectedProxyPorts(constants.DefaultHTTPPort, constants.DefaultHTTPSPort)),
			Entry("should deduplicate identical proxy ports",
				setProxies("http://proxy.example.com:8080", "https://proxy.example.com:8080"),
				expectedProxyPorts(8080)),
			Entry("should extract explicitly specified standard ports from proxy URLs",
				setProxies("http://proxy.example.com:80", "https://proxy.example.com:443"),
				expectedProxyPorts(constants.DefaultHTTPPort, constants.DefaultHTTPSPort)),
			Entry("should handle proxy URLs with authentication",
				setProxies("http://user:password@proxy.example.com:8080", "https://user:password@proxy.example.com:8443"),
				expectedProxyPorts(8080, 8443)),
			Entry("should handle IPv6 proxy URLs",
				setProxies("http://[::1]:8080", ""),
				expectedProxyPorts(8080)),
			Entry("should handle empty string proxy URLs gracefully",
				setProxies("", "https://proxy.example.com:8443"),
				expectedProxyPorts(8443)),
			Entry("should handle mix of uppercase and lowercase proxy environment variables",
				map[string]string{
					"HTTP_PROXY":  "http://proxy.example.com:8080",
					"https_proxy": "https://proxy.example.com:8443",
				},
				expectedProxyPorts(8080, 8443)),
		)

		DescribeTable("with unknown schemes and invalid proxy URLs", func(envVars map[string]string) {
			for key, value := range envVars {
				os.Setenv(key, value)
			}
			Expect(func() { GetProxyPorts() }).To(Panic())
		},
			Entry("should panic when proxy URLs have unknown schemes without ports",
				setProxies("proxy://proxy.example.com", "invalid://proxy.example.com")),
			Entry("should panic on malformed proxy URLs",
				setProxies("not-a-valid-url", "http://proxy.example.com:8080")),
			Entry("should panic with invalid ports on urls",
				setProxies("http://proxy.example.com:invalid", "https://proxy.example.com:no-a-port")),
		)

		It("should be empty portMap when no proxy environment variables are set", func() {
			portMap := GetProxyPorts()
			Expect(portMap).To(BeEmpty())
		})
	})
})

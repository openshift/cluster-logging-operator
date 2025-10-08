package network

import (
	"fmt"
	"net/url"
	"strconv"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/set"
)

// GetOutputPortsWithProtocols extracts all unique ports with their protocols from the given outputs.
// It parses URLs to extract ports and protocols, or uses default values based on the output type.
func GetOutputPortsWithProtocols(outputs []obs.OutputSpec) []factory.PortProtocol {
	portProtocolMap := map[factory.PortProtocol]bool{}

	for _, output := range outputs {
		portProtocols := getPortProtocolFromOutputURL(output)
		for _, pp := range portProtocols {
			if pp.Port > 0 {
				portProtocolMap[pp] = true
			}
		}
	}

	result := make([]factory.PortProtocol, 0, len(portProtocolMap))
	for pp := range portProtocolMap {
		result = append(result, pp)
	}

	return result
}

// GetInputPorts extracts all unique ports from the given input receiver specs.
// It returns the ports that input receivers are configured to listen on.
func GetInputPorts(inputs []obs.InputSpec) []int32 {
	portSet := set.New[int32]()

	for _, input := range inputs {
		if input.Type == obs.InputTypeReceiver && input.Receiver != nil {
			if input.Receiver.Port > 0 {
				portSet.Insert(input.Receiver.Port)
			}
		}
	}

	return portSet.UnsortedList()
}

// getPortProtocolFromOutputURL extracts all ports with protocols from an output spec's URL.
// For most outputs, it returns a single port with protocol. For Kafka, it returns ports from all brokers.
func getPortProtocolFromOutputURL(output obs.OutputSpec) []factory.PortProtocol {
	// Handle different output types
	var urlStr string
	switch output.Type {
	case obs.OutputTypeElasticsearch:
		if output.Elasticsearch != nil {
			urlStr = output.Elasticsearch.URL
		}
	case obs.OutputTypeSplunk:
		if output.Splunk != nil {
			urlStr = output.Splunk.URL
		}
	case obs.OutputTypeLoki:
		if output.Loki != nil {
			urlStr = output.Loki.URL
		}
	case obs.OutputTypeSyslog:
		if output.Syslog != nil {
			urlStr = output.Syslog.URL
		}
	case obs.OutputTypeOTLP:
		if output.OTLP != nil {
			urlStr = output.OTLP.URL
		}
	case obs.OutputTypeHTTP:
		if output.HTTP != nil {
			urlStr = output.HTTP.URL
		}
	case obs.OutputTypeCloudwatch:
		if output.Cloudwatch != nil && output.Cloudwatch.URL != "" {
			urlStr = output.Cloudwatch.URL
		}
	case obs.OutputTypeKafka:
		// For kafka, prefer the URL over the broker URLs
		if output.Kafka != nil {
			if output.Kafka.URL != "" {
				urlStr = output.Kafka.URL
			} else if len(output.Kafka.Brokers) > 0 {
				return getKafkaBrokerPortProtocols(output.Kafka.Brokers)
			}
		}
	}

	// Try to parse port from URL
	if urlStr != "" {
		if port := parsePortProtocolFromURL(urlStr); port != nil {
			return []factory.PortProtocol{{Port: port.Port, Protocol: port.Protocol}}
		}
	}

	// Fall back to default port with TCP protocol
	if defaultPort := getDefaultPort(output.Type, urlStr); defaultPort > 0 {
		return []factory.PortProtocol{{Port: defaultPort, Protocol: corev1.ProtocolTCP}}
	}

	return []factory.PortProtocol{}
}

// getKafkaBrokerPortProtocols extracts ports with protocols from all Kafka brokers.
func getKafkaBrokerPortProtocols(brokers []obs.BrokerURL) []factory.PortProtocol {
	portProtocolSlice := []factory.PortProtocol{}

	// Check all broker URLs
	for _, broker := range brokers {
		if port := parsePortProtocolFromURL(string(broker)); port != nil {
			portProtocolSlice = append(portProtocolSlice, factory.PortProtocol{Port: port.Port, Protocol: port.Protocol})
		} else {
			// If no port in broker URL, use default port based on scheme
			if defaultPort := getDefaultPort(obs.OutputTypeKafka, string(broker)); defaultPort > 0 {
				portProtocolSlice = append(portProtocolSlice, factory.PortProtocol{Port: defaultPort, Protocol: corev1.ProtocolTCP})
			}
		}
	}

	return portProtocolSlice
}

// parsePortProtocolFromURL extracts the port from a URL string.
// Returns nil if the port cannot be determined.
func parsePortProtocolFromURL(urlStr string) *factory.PortProtocol {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	// Determine protocol from scheme
	var protocol corev1.Protocol
	switch parsedURL.Scheme {
	case "udp":
		protocol = corev1.ProtocolUDP
	default:
		// Default to TCP
		protocol = corev1.ProtocolTCP
	}

	portStr := parsedURL.Port()
	if portStr != "" {
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err == nil && port > 0 {
			return &factory.PortProtocol{Port: int32(port), Protocol: protocol}
		}
	}

	return nil
}

// getDefaultPort returns the default port for a given output type based on the URL scheme or the default port for the output type.
func getDefaultPort(outputType obs.OutputType, urlStr string) int32 {
	// Parse URL to determine scheme for kafka and http/https to return appropriate default port
	var scheme string
	if urlStr != "" {
		if parsedURL, err := url.Parse(urlStr); err == nil {
			scheme = parsedURL.Scheme
		}
	}

	switch outputType {
	case obs.OutputTypeElasticsearch:
		return 9200
	case obs.OutputTypeSplunk:
		return 8088
	case obs.OutputTypeLoki:
		return 3100
	case obs.OutputTypeSyslog:
		return 514
	case obs.OutputTypeOTLP:
		return 4318
	case obs.OutputTypeLokiStack: // LokiStack uses 8080
		return 8080
	case obs.OutputTypeCloudwatch, obs.OutputTypeAzureMonitor, obs.OutputTypeGoogleCloudLogging:
		return 443
	case obs.OutputTypeKafka:
		// Kafka uses 9092 for plaintext (tcp), 9093 for TLS
		if scheme == "tls" {
			return 9093
		}
		return 9092
	case obs.OutputTypeHTTP:
		if scheme == "http" {
			return 80
		}
		return 443 // https
	}
	panic(fmt.Sprintf("unknown output type: %s", outputType))
}

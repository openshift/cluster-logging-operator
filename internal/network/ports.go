package network

import (
	"fmt"
	"net/url"
	"strconv"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetOutputPortsWithProtocols extracts all unique ports with their protocols from the given outputs.
// It parses URLs to extract ports and protocols, or uses default values based on the output type.
func GetOutputPortsWithProtocols(outputs []obs.OutputSpec, portProtocolMap map[factory.PortProtocol]bool) {
	for _, output := range outputs {
		portProtocols := getPortProtocolFromOutputURL(output)
		for _, pp := range portProtocols {
			if pp.Port > 0 {
				portProtocolMap[pp] = true
			}
		}
	}
}

// GetInputPorts extracts all unique ports from the given input receiver specs.
// It returns the ports that input receivers are configured to listen on.
func GetInputPorts(inputs []obs.InputSpec) []int32 {
	portSet := sets.NewInt32()

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
		if output.Cloudwatch != nil {
			urlStr = output.Cloudwatch.URL
		}
	case obs.OutputTypeS3:
		if output.S3 != nil {
			urlStr = output.S3.URL
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
	if defaultPort := getDefaultOutputPort(output.Type, urlStr); defaultPort > 0 {
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
			if defaultPort := getDefaultOutputPort(obs.OutputTypeKafka, string(broker)); defaultPort > 0 {
				portProtocolSlice = append(portProtocolSlice, factory.PortProtocol{Port: defaultPort, Protocol: corev1.ProtocolTCP})
			}
		}
	}

	return portProtocolSlice
}

func parsePortString(portStr string) (int32, bool) {
	if port, err := strconv.ParseInt(portStr, 10, 32); err == nil && port > 0 {
		return int32(port), true
	}
	return 0, false
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

	if port, ok := parsePortString(parsedURL.Port()); ok {
		return &factory.PortProtocol{Port: port, Protocol: protocol}
	}

	return nil
}

// getDefaultOutputPort returns the default port for a given output type based on the URL scheme or the default port for the output type.
func getDefaultOutputPort(outputType obs.OutputType, urlStr string) int32 {
	// Parse URL to determine scheme for kafka and http/https to return appropriate default port
	var scheme string
	if urlStr != "" {
		if parsedURL, err := url.Parse(urlStr); err == nil {
			scheme = parsedURL.Scheme
		}
	}

	if scheme == "https" {
		return 443
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
	case obs.OutputTypeCloudwatch, obs.OutputTypeAzureMonitor, obs.OutputTypeGoogleCloudLogging, obs.OutputTypeS3:
		return 443
	case obs.OutputTypeKafka:
		// Kafka uses 9092 for plaintext (tcp), 9093 for TLS
		if scheme == "tls" {
			return 9093
		}
		return 9092
	case obs.OutputTypeHTTP:
		return 80
	}
	panic(fmt.Sprintf("unknown output type: %s", outputType))
}

// getDefaultProxyPort returns the default port for a given proxy environment variable based on the URL scheme.
func getDefaultProxyPort(scheme string) (int32, bool) {
	switch scheme {
	case "http":
		return 80, true
	case "https":
		return 443, true
	}
	return 0, false
}

// GetProxyPorts extracts unique ports from cluster-wide proxy environment variables.
// It parses HTTP_PROXY and HTTPS_PROXY URLs to determine explicit proxy ports,
// or adds default ports (80 for HTTP, 443 for HTTPS) when no port is specified.
func GetProxyPorts(portProtocolMap map[factory.PortProtocol]bool) {
	// Get proxy environment variables and parse them for additional explicit ports
	proxyEnvVars := utils.GetProxyEnvVars()

	for _, envVar := range proxyEnvVars {
		// Skip non-proxy environment variables or empty values
		if (envVar.Name != "http_proxy" && envVar.Name != "https_proxy") || envVar.Value == "" {
			continue
		}

		// Parse URL for port extraction or default port determination
		parsedURL, err := url.Parse(envVar.Value)
		if err != nil {
			log.V(0).Error(err, "Failed to parse proxy URL", "url", envVar.Value)
			continue
		}

		var port int32
		var ok bool
		// Extract port from URL or use default port based on URL scheme
		if port, ok = parsePortString(parsedURL.Port()); !ok {
			port, ok = getDefaultProxyPort(parsedURL.Scheme)
		}
		if ok {
			portProtocolMap[factory.PortProtocol{Port: port, Protocol: corev1.ProtocolTCP}] = true
		}
	}
}

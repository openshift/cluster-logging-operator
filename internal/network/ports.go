package network

import (
	"fmt"
	"net/url"
	"strconv"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var defaultHTTPSTCPPort = factory.PortProtocol{Port: constants.DefaultHTTPSPort, Protocol: corev1.ProtocolTCP}

// DetermineEgressPortProtocols determines the egress ports needed based on outputs and policy rule set.
// Returns collected ports from outputs + proxy configuration for RestrictIngressEgress.
func DetermineEgressPortProtocols(outputs []obs.OutputSpec, policyRuleSet obs.NetworkPolicyRuleSetType) []factory.PortProtocol {
	if policyRuleSet != obs.NetworkPolicyRuleSetTypeRestrictIngressEgress {
		return nil
	}

	// Collect output ports and proxy ports
	egressPortMap := GetOutputPortsWithProtocols(outputs)
	for pp := range GetProxyPorts() {
		egressPortMap[pp] = true
	}

	// Convert map to slice
	egressPorts := make([]factory.PortProtocol, 0, len(egressPortMap))
	for pp := range egressPortMap {
		egressPorts = append(egressPorts, pp)
	}
	return egressPorts
}

// DetermineIngressPortProtocols determines the ingress ports needed based on inputs and policy rule set.
// Returns ports from receiver inputs for RestrictIngressEgress.
func DetermineIngressPortProtocols(inputs []obs.InputSpec, policyRuleSet obs.NetworkPolicyRuleSetType) []int32 {
	if policyRuleSet != obs.NetworkPolicyRuleSetTypeRestrictIngressEgress {
		return nil
	}
	return GetInputPorts(inputs)
}

// GetOutputPortsWithProtocols extracts all unique ports with their protocols from the given outputs based on their URL(s).
// It parses URLs to extract ports and protocols.
func GetOutputPortsWithProtocols(outputs []obs.OutputSpec) map[factory.PortProtocol]bool {
	var portProtocols []factory.PortProtocol
	for _, output := range outputs {
		portProtocols = append(portProtocols, getPortProtocolFromOutputURLs(output)...)
	}

	portProtocolMap := make(map[factory.PortProtocol]bool, len(portProtocols))
	for _, pp := range portProtocols {
		if pp.Port > 0 {
			portProtocolMap[pp] = true
		}
	}
	return portProtocolMap
}

// GetInputPorts extracts all unique ports from the given input receiver specs.
// It returns the ports that input receivers are configured to listen on.
func GetInputPorts(inputs []obs.InputSpec) []int32 {
	portSet := sets.NewInt32()

	for _, input := range inputs {
		if input.Type == obs.InputTypeReceiver && input.Receiver != nil && input.Receiver.Port > 0 {
			portSet.Insert(input.Receiver.Port)
		}
	}

	return portSet.UnsortedList()
}

// getPortProtocolFromOutputURLs extracts all ports with protocols from an output spec's URL.
// For most outputs, it returns a slice with a single port protocol.
// For Kafka, it returns ports from all brokers or the URL if provided.
// For HTTP, it returns ports from the URL and proxy URL if provided.
// Returns port 443 for Google Cloud Logging and Azure Monitor as well as Cloudwatch and S3 if no URL is provided.
func getPortProtocolFromOutputURLs(output obs.OutputSpec) []factory.PortProtocol {
	// Gather all URL strings from the output spec
	var urlSlice []string

	switch output.Type {
	case obs.OutputTypeElasticsearch:
		if output.Elasticsearch != nil {
			urlSlice = append(urlSlice, output.Elasticsearch.URL)
		}
	case obs.OutputTypeSplunk:
		if output.Splunk != nil {
			urlSlice = append(urlSlice, output.Splunk.URL)
		}
	case obs.OutputTypeLoki:
		if output.Loki != nil {
			urlSlice = append(urlSlice, output.Loki.URL, output.Loki.ProxyURL)
		}
	case obs.OutputTypeSyslog:
		if output.Syslog != nil {
			urlSlice = append(urlSlice, output.Syslog.URL)
		}
	case obs.OutputTypeOTLP:
		if output.OTLP != nil {
			urlSlice = append(urlSlice, output.OTLP.URL)
		}
	case obs.OutputTypeHTTP:
		if output.HTTP != nil {
			urlSlice = append(urlSlice, output.HTTP.URL, output.HTTP.ProxyURL)
		}
	case obs.OutputTypeKafka:
		if output.Kafka != nil {
			urlSlice = append(urlSlice, getKafkaAndBrokerURLs(*output.Kafka)...)
		}
	case obs.OutputTypeCloudwatch:
		if output.Cloudwatch == nil {
			return nil
		}
		// Cloudwatch URL is optional; default to HTTPS port 443 if not specified
		if output.Cloudwatch.URL == "" {
			return []factory.PortProtocol{defaultHTTPSTCPPort}
		}
		urlSlice = append(urlSlice, output.Cloudwatch.URL)
	case obs.OutputTypeS3:
		if output.S3 == nil {
			return nil
		}
		// S3 URL is optional; default to HTTPS port 443 when not specified
		if output.S3.URL == "" {
			return []factory.PortProtocol{defaultHTTPSTCPPort}
		}
		urlSlice = append(urlSlice, output.S3.URL)
		// LokiStack internal port is 8080 for both HTTP and OTLP
	case obs.OutputTypeLokiStack:
		return []factory.PortProtocol{{Port: 8080, Protocol: corev1.ProtocolTCP}}
	case obs.OutputTypeGoogleCloudLogging, obs.OutputTypeAzureMonitor:
		// GCL and Azure Monitor don't have URL fields and use the default HTTPS port
		return []factory.PortProtocol{defaultHTTPSTCPPort}
	default:
		panic(fmt.Sprintf("Unsupported output type: %s", output.Type))
	}

	// Extract ports from output URL(s) if any
	portProtocolSlice := make([]factory.PortProtocol, 0, len(urlSlice))
	for _, url := range urlSlice {
		if port := parsePortProtocolFromURL(url); port != nil {
			portProtocolSlice = append(portProtocolSlice, *port)
		}
	}

	return portProtocolSlice
}

// getKafkaAndBrokerURLs extracts the URL and broker URLs from a Kafka output spec.
func getKafkaAndBrokerURLs(kafka obs.Kafka) []string {
	urls := []string{kafka.URL}

	for _, broker := range kafka.Brokers {
		urls = append(urls, string(broker))
	}

	return urls
}

// parsePortString extracts the port from a string if it is a valid integer.
func parsePortString(portStr string) (int32, bool) {
	if port, err := strconv.ParseInt(portStr, 10, 32); err == nil && port > 0 {
		return int32(port), true
	}
	return 0, false
}

// parsePortProtocolFromURL extracts the port from a URL string.
// Defaults the ports to 80 for HTTP, 443 for HTTPS if no port is specified.
// Non-standard schemes (UDP, TLS, TCP) should always have a port associated with its URL and will panic if no port is specified.
func parsePortProtocolFromURL(urlStr string) *factory.PortProtocol {
	if urlStr == "" {
		return nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL %q for port and protocol: %v", urlStr, err))
	}

	// Determine port and protocol from scheme
	var port int32
	var protocol corev1.Protocol

	switch parsedURL.Scheme {
	case "udp":
		protocol = corev1.ProtocolUDP
	case "http":
		port = constants.DefaultHTTPPort
		protocol = corev1.ProtocolTCP
	default:
		port = constants.DefaultHTTPSPort
		protocol = corev1.ProtocolTCP
	}

	if parsedPort, ok := parsePortString(parsedURL.Port()); ok && parsedPort != 0 {
		port = parsedPort
		// Panic on non-standard schemes without a port
	} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		panic(fmt.Sprintf("non-standard scheme %q requires an explicit port in URL %q", parsedURL.Scheme, urlStr))
	}

	return &factory.PortProtocol{Port: port, Protocol: protocol}
}

// GetProxyPorts extracts unique ports from HTTP_PROXY and HTTPS_PROXY environment variables by parsing their URLs
func GetProxyPorts() map[factory.PortProtocol]bool {
	// Get proxy environment variables and parse them for additional explicit ports
	proxyEnvVars := utils.GetProxyEnvVars()

	proxyPortProtocolMap := make(map[factory.PortProtocol]bool, len(proxyEnvVars))
	for _, envVar := range proxyEnvVars {
		// Skip NO_PROXY and empty values
		if envVar.Name == "no_proxy" || envVar.Value == "" {
			continue
		}

		if portProtocol := parsePortProtocolFromURL(envVar.Value); portProtocol != nil {
			proxyPortProtocolMap[*portProtocol] = true
		}
	}
	return proxyPortProtocolMap
}

package tls

import "time"

// ScanResult represents a single TLS scan result
type ScanResult struct {
	IP           string
	Port         string
	Protocol     string
	Service      string
	Pod          string
	Namespace    string
	Status       string
	TLSVersions  []string
	Ciphers      string
	Component    string
	ListenAddr   string
	TLSReadiness *TLSReadiness
}

// ScanOutput represents the top-level JSON output from the TLS scanner
type ScanOutput struct {
	Timestamp  time.Time  `json:"timestamp"`
	TotalIPs   int        `json:"total_ips"`
	ScannedIPs int        `json:"scanned_ips"`
	IPResults  []IPResult `json:"ip_results"`
}

// IPResult represents scan results for a single IP address
type IPResult struct {
	IP                 string             `json:"ip"`
	Status             string             `json:"status"`
	OpenPorts          []int              `json:"open_ports"`
	PortResults        []PortResult       `json:"port_results"`
	OpenShiftComponent OpenShiftComponent `json:"openshift_component"`
}

type OpenShiftComponent struct {
	Component           string `json:"component"`
	SourceLocation      string `json:"source_location"`
	MaintainerComponent string `json:"maintainer_component"`
	IsBundler           bool   `json:"is_bundle"`
}

// PortResult represents scan results for a single port
type PortResult struct {
	Port         int           `json:"port"`
	Protocol     string        `json:"protocol"`
	State        string        `json:"state"`
	Service      string        `json:"service"`
	Status       string        `json:"status"`
	Reason       string        `json:"reason"`
	TLSReadiness *TLSReadiness `json:"tls_readiness,omitempty"`
}

// TLSReadiness represents TLS capability information for a port
type TLSReadiness struct {
	TLS13Offered bool   `json:"tls13_offered"`
	TLS12Only    bool   `json:"tls12_only"`
	PQCCapable   bool   `json:"pqc_capable"`
	Notes        string `json:"notes"`
}

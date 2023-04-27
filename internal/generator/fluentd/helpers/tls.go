package helpers

import (
	"strings"
)

func TLSMinVersion(tlsProfileSpecVersion string) string {
	switch tlsProfileSpecVersion {
	case "VersionTLS10":
		return "TLS1_1" // no TLS1_0 in fluentd conf
	case "VersionTLS11":
		return "TLS1_1"
	case "VersionTLS12":
		return "TLS1_2"
	case "VersionTLS13":
		return "TLS1_3"
	}
	return ""
}

func TLSCiphers(ciphers []string) string {
	return strings.Join(ciphers, ":")
}

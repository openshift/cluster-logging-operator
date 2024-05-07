package helpers

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"strings"
)

var (

	// FrameworkOptionWithDefaultTLSCiphers is a reusable test helper with default minTLS and Ciphers
	FrameworkOptionWithDefaultTLSCiphers = framework.Options{
		framework.MinTLSVersion: string(tls.DefaultMinTLSVersion),
		framework.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
	}
)

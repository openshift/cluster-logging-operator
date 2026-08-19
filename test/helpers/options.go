package helpers

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

var (

	// FrameworkOptionWithDefaultTLSCiphers is a reusable test helper with default minTLS, Ciphers, and Groups
	FrameworkOptionWithDefaultTLSCiphers = framework.Options{
		framework.MinTLSVersion: string(tls.DefaultMinTLSVersion),
		framework.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
		framework.Groups:        tls.TLSGroupsToOpenSSL(configv1.TLSProfiles[tls.DefaultTLSProfileType].Groups),
	}
)

package framework

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// TLSProfileInfo returns the minTLSVersion, ciphers, and groups given the available TLSSecurityProfile.
// Ciphers are returned as a delimited list using the provided separator.
// Groups are returned as a colon-separated string of OpenSSL curve names for Vector.
func TLSProfileInfo(op utils.Options, tlsSpec internalobs.TransportLayerSecurity, separator string) (string, string, string) {
	var tlsProfileSpec configv1.TLSProfileSpec
	if tlsSpec != nil && tlsSpec.GetTlsSecurityProfile() != nil {
		tlsProfileSpec = tls.GetClusterTLSProfileSpec(tlsSpec.GetTlsSecurityProfile())
	} else if _, ok := op[ClusterTLSProfileSpec]; ok {
		clusterSpec := op[ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		tlsProfileSpec = clusterSpec
	}

	minTlsVersion := tls.MinTLSVersion(tlsProfileSpec)
	cipherSuites := strings.Join(tls.TLSCiphers(tlsProfileSpec), separator)
	groups := tls.TLSGroupsToOpenSSL(tlsProfileSpec.Groups)
	if groups == "" {
		groups = tls.TLSGroupsToOpenSSL(configv1.TLSProfiles[tls.DefaultTLSProfileType].Groups)
	}
	return minTlsVersion, cipherSuites, groups
}

// SetTLSProfileOptionsFrom updates options to set the TLS profile based upon the output spec
func SetTLSProfileOptionsFrom(op utils.Options, tlsSpec internalobs.TransportLayerSecurity) {
	op[MinTLSVersion], op[Ciphers], op[Groups] = TLSProfileInfo(op, tlsSpec, ",")
}

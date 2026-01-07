package framework

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// TLSProfileInfo returns the minTLSVersion, ciphers as a delimited list given the available TLSSecurityProfile
func TLSProfileInfo(op utils.Options, tlsSpec internalobs.TransportLayerSecurity, separator string) (string, string) {
	var tlsProfileSpec configv1.TLSProfileSpec
	if tlsSpec != nil && tlsSpec.GetTlsSecurityProfile() != nil {
		tlsProfileSpec = tls.GetClusterTLSProfileSpec(tlsSpec.GetTlsSecurityProfile())
	} else if _, ok := op[ClusterTLSProfileSpec]; ok {
		clusterSpec := op[ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		tlsProfileSpec = clusterSpec
	}

	minTlsVersion := tls.MinTLSVersion(tlsProfileSpec)
	cipherSuites := strings.Join(tls.TLSCiphers(tlsProfileSpec), separator)
	return minTlsVersion, cipherSuites
}

// SetTLSProfileOptionsFrom updates options to set the TLS profile based upon the output spec
func SetTLSProfileOptionsFrom(op utils.Options, tlsSpec internalobs.TransportLayerSecurity) {
	op[MinTLSVersion], op[Ciphers] = TLSProfileInfo(op, tlsSpec, ",")
}

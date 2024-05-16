package framework

import (
	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"strings"
)

// TLSProfileInfo returns the minTLSVersion, ciphers as a delimited list given the available TLSSecurityProfile
func (op Options) TLSProfileInfo(outputSpec obs.OutputSpec, separator string) (string, string) {
	var tlsProfileSpec configv1.TLSProfileSpec
	if outputSpec.TLS != nil && outputSpec.TLS.TLSSecurityProfile != nil {
		tlsProfileSpec = tls.GetClusterTLSProfileSpec(outputSpec.TLS.TLSSecurityProfile)
	} else if _, ok := op[ClusterTLSProfileSpec]; ok {
		clusterSpec := op[ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		tlsProfileSpec = clusterSpec
	}

	minTlsVersion := tls.MinTLSVersion(tlsProfileSpec)
	cipherSuites := strings.Join(tls.TLSCiphers(tlsProfileSpec), separator)
	return minTlsVersion, cipherSuites
}

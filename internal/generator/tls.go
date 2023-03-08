package generator

import (
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"strings"
)

// TLSProfileInfo returns the minTLSVersion, ciphers as a comma-separated list given the available TLSSecurityProfile
func (op Options) TLSProfileInfo(clfProfile *configv1.TLSSecurityProfile, outputSpec logging.OutputSpec) (string, string) {
	var tlsProfileSpec *configv1.TLSProfileSpec
	if outputSpec.TLS != nil && outputSpec.TLS.TLSSecurityProfile != nil {
		tlsProfileSpec = configv1.TLSProfiles[outputSpec.TLS.TLSSecurityProfile.Type]
		if configv1.TLSProfileCustomType == outputSpec.TLS.TLSSecurityProfile.Type {
			minTLSVersion := outputSpec.TLS.TLSSecurityProfile.Custom.MinTLSVersion
			ciphers := outputSpec.TLS.TLSSecurityProfile.Custom.Ciphers
			return string(minTLSVersion), strings.Join(ciphers, ",")
		}
	} else if clfProfile != nil {
		tlsProfileSpec = configv1.TLSProfiles[clfProfile.Type]
		if configv1.TLSProfileCustomType == clfProfile.Type {
			minTLSVersion := clfProfile.Custom.MinTLSVersion
			ciphers := clfProfile.Custom.Ciphers
			return string(minTLSVersion), strings.Join(ciphers, ",")
		}
	} else if _, ok := op[ClusterTLSProfileSpec]; ok {
		clusterSpec := op[ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		tlsProfileSpec = &clusterSpec
	}
	if tlsProfileSpec == nil {
		return "", ""
	}

	minTlsVersion := tls.MinTLSVersion(*tlsProfileSpec)
	cipherSuites := strings.Join(tls.TLSCiphers(*tlsProfileSpec), ",")
	return minTlsVersion, cipherSuites
}

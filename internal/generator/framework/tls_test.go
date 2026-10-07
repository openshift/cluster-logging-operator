package framework_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

var _ = Describe("Options#TLSProfileInfo", func() {

	var (
		options = Options{}
	)

	Context("when a cluster profile is absent", func() {

		It("should use the defaults when clf profile is nil and TLS spec is nil", func() {
			minTLS, ciphers, groups := TLSProfileInfo(options, nil, ",")
			Expect(minTLS).To(BeEquivalentTo(tls.DefaultMinTLSVersion))
			Expect(ciphers).To(Equal(strings.Join(tls.DefaultTLSCiphers, ",")))
			Expect(groups).ToNot(BeEmpty())
		})
	})

	Context("when a cluster profile exists", func() {

		var (
			clusterCiphers       = []string{"a", "b", "c"}
			clusterMinTLSVersion = configv1.VersionTLS12
			outputProfile        *configv1.TLSSecurityProfile
			outputSpec           obs.OutputSpec
		)
		BeforeEach(func() {
			options = Options{}
			options[ClusterTLSProfileSpec] = configv1.TLSProfileSpec{
				Ciphers:       clusterCiphers,
				MinTLSVersion: clusterMinTLSVersion,
			}
			outputProfile = &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileOldType,
			}
			outputSpec = obs.OutputSpec{
				TLS: &obs.OutputTLSSpec{
					TLSSecurityProfile: outputProfile,
				},
			}
		})

		It("should prefer the output profile over the cluster profile", func() {
			minTLS, ciphers, groups := TLSProfileInfo(options, adapters.NewOutput(outputSpec), ",")
			// Old profile has TLS 1.0, but it gets upgraded to TLS 1.2 (CWE-327 hardening)
			Expect(minTLS).To(BeEquivalentTo(configv1.VersionTLS12))
			// Old profile contains insecure ciphers (DHE, CBC-mode); only secure AEAD+ECDHE ciphers are kept
			// Expected: 3 TLS 1.3 ciphers + 6 TLS 1.2 ECDHE-AEAD ciphers = 9 total (DHE/CBC filtered out)
			expectedCiphers := []string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
				"TLS_CHACHA20_POLY1305_SHA256",
				"ECDHE-ECDSA-AES128-GCM-SHA256",
				"ECDHE-RSA-AES128-GCM-SHA256",
				"ECDHE-ECDSA-AES256-GCM-SHA384",
				"ECDHE-RSA-AES256-GCM-SHA384",
				"ECDHE-ECDSA-CHACHA20-POLY1305",
				"ECDHE-RSA-CHACHA20-POLY1305",
			}
			Expect(ciphers).To(Equal(strings.Join(expectedCiphers, ",")))
			Expect(groups).ToNot(BeEmpty())
		})

		It("should prefer the cluster profile when the forwarder and TLS spec are nil", func() {
			minTLS, ciphers, _ := TLSProfileInfo(options, nil, ",")
			Expect(minTLS).To(BeEquivalentTo(clusterMinTLSVersion))
			// Invalid ciphers ["a", "b", "c"] are filtered out, defaults are returned (CWE-327 hardening)
			Expect(ciphers).To(Equal(strings.Join(tls.DefaultTLSCiphers, ",")))
		})
	})

})

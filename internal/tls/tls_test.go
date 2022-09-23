package tls

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	TLSProfileModern = `
config_diagnostics = 1

openssl_conf = default_conf_section

[default_conf_section]
ssl_conf = ssl_section

[ssl_section]
system_default = system_default_section

[system_default_section]
MinProtocol = TLSv1.3
CipherSuites = TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256
`
	TLSProfileIntermediate = `
config_diagnostics = 1

openssl_conf = default_conf_section

[default_conf_section]
ssl_conf = ssl_section

[ssl_section]
system_default = system_default_section

[system_default_section]
MinProtocol = TLSv1.2
CipherSuites = TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256, ECDHE-ECDSA-AES128-GCM-SHA256, ECDHE-RSA-AES128-GCM-SHA256, ECDHE-ECDSA-AES256-GCM-SHA384, ECDHE-RSA-AES256-GCM-SHA384, ECDHE-ECDSA-CHACHA20-POLY1305, ECDHE-RSA-CHACHA20-POLY1305, DHE-RSA-AES128-GCM-SHA256, DHE-RSA-AES256-GCM-SHA384
`
	TLSProfileOldCompatibility = `
config_diagnostics = 1

openssl_conf = default_conf_section

[default_conf_section]
ssl_conf = ssl_section

[ssl_section]
system_default = system_default_section

[system_default_section]
MinProtocol = TLSv1.0
CipherSuites = TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256, ECDHE-ECDSA-AES128-GCM-SHA256, ECDHE-RSA-AES128-GCM-SHA256, ECDHE-ECDSA-AES256-GCM-SHA384, ECDHE-RSA-AES256-GCM-SHA384, ECDHE-ECDSA-CHACHA20-POLY1305, ECDHE-RSA-CHACHA20-POLY1305, DHE-RSA-AES128-GCM-SHA256, DHE-RSA-AES256-GCM-SHA384, DHE-RSA-CHACHA20-POLY1305, ECDHE-ECDSA-AES128-SHA256, ECDHE-RSA-AES128-SHA256, ECDHE-ECDSA-AES128-SHA, ECDHE-RSA-AES128-SHA, ECDHE-ECDSA-AES256-SHA384, ECDHE-RSA-AES256-SHA384, ECDHE-ECDSA-AES256-SHA, ECDHE-RSA-AES256-SHA, DHE-RSA-AES128-SHA256, DHE-RSA-AES256-SHA256, AES128-GCM-SHA256, AES256-GCM-SHA384, AES128-SHA256, AES256-SHA256, AES128-SHA, AES256-SHA, DES-CBC3-SHA
`
)

var _ = Describe("Based on tlsSecurityProfile", func() {
	Context("should generate OPENSSL_CONF for", func() {
		It("Modern profile", func() {
			c, err := opensslConf(*configv1.TLSProfiles[configv1.TLSProfileModernType])
			Expect(err).To(BeNil())
			Expect(strings.TrimSpace(c)).To(matchers.EqualTrimLines(TLSProfileModern))
		})
		It("Intermediate profile", func() {
			c, err := opensslConf(*configv1.TLSProfiles[configv1.TLSProfileIntermediateType])
			Expect(err).To(BeNil())
			Expect(strings.TrimSpace(c)).To(matchers.EqualTrimLines(TLSProfileIntermediate))
		})
		It("Old profile", func() {
			c, err := opensslConf(*configv1.TLSProfiles[configv1.TLSProfileOldType])
			Expect(err).To(BeNil())
			Expect(strings.TrimSpace(c)).To(matchers.EqualTrimLines(TLSProfileOldCompatibility))
		})
		It("Custom profile", func() {
			customProfile := `
config_diagnostics = 1

openssl_conf = default_conf_section

[default_conf_section]
ssl_conf = ssl_section

[ssl_section]
system_default = system_default_section

[system_default_section]
MinProtocol = TLSv1.2
CipherSuites = TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384
`
			customTLSSpec := configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers:       []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
			}
			c, err := opensslConf(customTLSSpec)
			Expect(err).To(BeNil())
			Expect(strings.TrimSpace(c)).To(matchers.EqualTrimLines(customProfile))
		})
	})
})

func TestTLSSecurityProfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLS Security Profile")
}

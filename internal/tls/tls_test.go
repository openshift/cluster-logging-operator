package tls_test

import (
	"crypto/tls"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	. "github.com/openshift/cluster-logging-operator/internal/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#TLSCiphers", func() {
	It("should return the default ciphers when none are defined", func() {
		Expect(TLSCiphers(configv1.TLSProfileSpec{})).To(BeEquivalentTo(DefaultTLSCiphers))
	})
	It("should filter out insecure ciphers (CWE-327)", func() {
		insecureCiphers := []string{
			"TLS_AES_128_GCM_SHA256", // Secure - should be kept
			"DES-CBC3-SHA",           // Insecure - should be filtered
			"AES128-SHA",             // Insecure - should be filtered
			"ECDHE-RSA-AES128-GCM-SHA256", // Secure - should be kept
		}
		result := TLSCiphers(configv1.TLSProfileSpec{Ciphers: insecureCiphers})
		Expect(result).To(ContainElement("TLS_AES_128_GCM_SHA256"))
		Expect(result).To(ContainElement("ECDHE-RSA-AES128-GCM-SHA256"))
		Expect(result).ToNot(ContainElement("DES-CBC3-SHA"))
		Expect(result).ToNot(ContainElement("AES128-SHA"))
	})
	It("should return defaults when all ciphers are insecure", func() {
		insecureCiphers := []string{"DES-CBC3-SHA", "AES128-SHA", "AES256-SHA"}
		result := TLSCiphers(configv1.TLSProfileSpec{Ciphers: insecureCiphers})
		Expect(result).To(Equal(DefaultTLSCiphers))
	})
	It("should accept all ciphers from the Intermediate profile", func() {
		intermediateProfile := *configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
		result := TLSCiphers(intermediateProfile)
		Expect(len(result)).To(Equal(len(intermediateProfile.Ciphers)))
		for _, cipher := range intermediateProfile.Ciphers {
			Expect(result).To(ContainElement(cipher))
		}
	})
	It("should filter insecure ciphers from the Old profile", func() {
		oldProfile := *configv1.TLSProfiles[configv1.TLSProfileOldType]
		result := TLSCiphers(oldProfile)
		// Should have fewer ciphers than the Old profile (insecure ones filtered)
		Expect(len(result)).To(BeNumerically("<", len(oldProfile.Ciphers)))
		// Should not contain 3DES
		Expect(result).ToNot(ContainElement("DES-CBC3-SHA"))
		// Should not contain AES-SHA (weak MAC)
		Expect(result).ToNot(ContainElement("AES128-SHA"))
		Expect(result).ToNot(ContainElement("AES256-SHA"))
	})
})

var _ = Describe("#MinTLSVersion", func() {
	It("should return the default min TLS version when not defined", func() {
		Expect(string(DefaultMinTLSVersion)).To(Equal(MinTLSVersion(configv1.TLSProfileSpec{})))
	})
	It("should return the profile min TLS version when defined", func() {
		Expect(string(configv1.VersionTLS13)).To(Equal(MinTLSVersion(configv1.TLSProfileSpec{MinTLSVersion: configv1.VersionTLS13})))
	})
	It("should reject TLS 1.0 and return TLS 1.2 (CWE-327)", func() {
		Expect(MinTLSVersion(configv1.TLSProfileSpec{MinTLSVersion: configv1.VersionTLS10})).
			To(Equal(string(configv1.VersionTLS12)))
	})
	It("should reject TLS 1.1 and return TLS 1.2 (CWE-327)", func() {
		Expect(MinTLSVersion(configv1.TLSProfileSpec{MinTLSVersion: configv1.VersionTLS11})).
			To(Equal(string(configv1.VersionTLS12)))
	})
	It("should accept TLS 1.2", func() {
		Expect(MinTLSVersion(configv1.TLSProfileSpec{MinTLSVersion: configv1.VersionTLS12})).
			To(Equal(string(configv1.VersionTLS12)))
	})
})

var _ = Describe("#TLSGroups", func() {
	It("should return the default groups when none are in the profile", func() {
		Expect(TLSGroups(configv1.TLSProfileSpec{})).To(Equal(DefaultTLSGroups))
	})
	It("should return the profile groups when they are defined", func() {
		profile := configv1.TLSProfileSpec{
			Groups: []configv1.TLSGroup{configv1.TLSGroupX25519, configv1.TLSGroupSecP256r1},
		}
		Expect(TLSGroups(profile)).To(Equal([]string{"X25519", "secp256r1"}))
	})
})

var _ = Describe("#TLSGroupToID", func() {
	It("should convert valid group names to CurveID", func() {
		id, err := TLSGroupToID(configv1.TLSGroupX25519)
		Expect(err).ToNot(HaveOccurred())
		Expect(id).To(Equal(tls.X25519))

		id, err = TLSGroupToID(configv1.TLSGroupSecP256r1)
		Expect(err).ToNot(HaveOccurred())
		Expect(id).To(Equal(tls.CurveP256))

		id, err = TLSGroupToID(configv1.TLSGroupX25519MLKEM768)
		Expect(err).ToNot(HaveOccurred())
		Expect(id).To(Equal(tls.X25519MLKEM768))
	})

	It("should return error for unsupported group names", func() {
		_, err := TLSGroupToID(configv1.TLSGroup("INVALID"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported TLS group"))
	})
})

var _ = Describe("#TLSGroupsToOpenSSL", func() {
	It("should convert groups to colon-separated OpenSSL names", func() {
		groups := []configv1.TLSGroup{
			configv1.TLSGroupX25519,
			configv1.TLSGroupSecP256r1,
			configv1.TLSGroupSecP384r1,
		}
		Expect(TLSGroupsToOpenSSL(groups)).To(Equal("X25519:prime256v1:secp384r1"))
	})

	It("should return empty string for nil groups", func() {
		Expect(TLSGroupsToOpenSSL(nil)).To(Equal(""))
	})

	It("should skip unknown groups", func() {
		groups := []configv1.TLSGroup{
			configv1.TLSGroupX25519,
			configv1.TLSGroup("UNKNOWN"),
			configv1.TLSGroupSecP384r1,
		}
		Expect(TLSGroupsToOpenSSL(groups)).To(Equal("X25519:secp384r1"))
	})
})
var _ = Describe("#CipherSuiteStringToID", func() {
	It("should convert valid cipher suite names to IDs", func() {
		// Test a known cipher suite
		id, err := CipherSuiteStringToID("TLS_AES_128_GCM_SHA256")
		Expect(err).ToNot(HaveOccurred())
		Expect(id).To(Equal(tls.TLS_AES_128_GCM_SHA256))
	})

	It("should return error for invalid cipher suite names", func() {
		_, err := CipherSuiteStringToID("INVALID_CIPHER")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported cipher suite"))
	})
})

var _ = Describe("#TLSVersionToConstant", func() {
	It("should convert TLS version strings to constants", func() {
		version, err := TLSVersionToConstant(configv1.VersionTLS12)
		Expect(err).ToNot(HaveOccurred())
		Expect(version).To(Equal(uint16(tls.VersionTLS12)))

		version, err = TLSVersionToConstant(configv1.VersionTLS13)
		Expect(err).ToNot(HaveOccurred())
		Expect(version).To(Equal(uint16(tls.VersionTLS13)))
	})

	It("should return default for unknown versions", func() {
		version, err := TLSVersionToConstant("")
		Expect(err).ToNot(HaveOccurred())
		Expect(version).To(Equal(uint16(tls.VersionTLS12)))
	})
})

var _ = Describe("#TLSConfigFromProfile", func() {
	It("should create TLS config with default values when profile is empty", func() {
		config, err := TLSConfigFromProfile(configv1.TLSProfileSpec{})
		Expect(err).ToNot(HaveOccurred())
		Expect(config.MinVersion).To(Equal(uint16(tls.VersionTLS12)))
	})

	It("should create TLS config with specified min version", func() {
		config, err := TLSConfigFromProfile(configv1.TLSProfileSpec{
			MinTLSVersion: configv1.VersionTLS13,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(config.MinVersion).To(Equal(uint16(tls.VersionTLS13)))
	})

	It("should create TLS config with specified cipher suites", func() {
		config, err := TLSConfigFromProfile(configv1.TLSProfileSpec{
			Ciphers: []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(config.CipherSuites).To(HaveLen(2))
		Expect(config.CipherSuites).To(ContainElement(tls.TLS_AES_128_GCM_SHA256))
		Expect(config.CipherSuites).To(ContainElement(tls.TLS_AES_256_GCM_SHA384))
	})

	It("should skip invalid cipher suites", func() {
		config, err := TLSConfigFromProfile(configv1.TLSProfileSpec{
			Ciphers: []string{"TLS_AES_128_GCM_SHA256", "INVALID_CIPHER", "TLS_AES_256_GCM_SHA384"},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(config.CipherSuites).To(HaveLen(2))
		Expect(config.CipherSuites).To(ContainElement(tls.TLS_AES_128_GCM_SHA256))
		Expect(config.CipherSuites).To(ContainElement(tls.TLS_AES_256_GCM_SHA384))
	})

	It("should set CurvePreferences from profile groups", func() {
		config, err := TLSConfigFromProfile(configv1.TLSProfileSpec{
			Groups: []configv1.TLSGroup{configv1.TLSGroupX25519, configv1.TLSGroupSecP256r1},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(config.CurvePreferences).To(HaveLen(2))
		Expect(config.CurvePreferences).To(ContainElement(tls.X25519))
		Expect(config.CurvePreferences).To(ContainElement(tls.CurveP256))
	})

	It("should skip unsupported groups", func() {
		config, err := TLSConfigFromProfile(configv1.TLSProfileSpec{
			Groups: []configv1.TLSGroup{configv1.TLSGroupX25519, "UNSUPPORTED", configv1.TLSGroupSecP384r1},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(config.CurvePreferences).To(HaveLen(2))
		Expect(config.CurvePreferences).To(ContainElement(tls.X25519))
		Expect(config.CurvePreferences).To(ContainElement(tls.CurveP384))
	})
})

var _ = Describe("isClusterAPIServer predicate", func() {
	It("should return true for cluster APIServer", func() {
		apiServer := &configv1.APIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: APIServerName,
			},
		}
		Expect(IsClusterAPIServer(apiServer)).To(BeTrue())
	})

	It("should return false for non-cluster APIServer", func() {
		apiServer := &configv1.APIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-cluster",
			},
		}
		Expect(IsClusterAPIServer(apiServer)).To(BeFalse())
	})
})

var _ = Describe("#CipherSuiteStringToID security tests (CWE-327)", func() {
	It("should reject insecure cipher suites", func() {
		insecureSuites := tls.InsecureCipherSuites()
		for _, suite := range insecureSuites {
			_, err := CipherSuiteStringToID(suite.Name)
			Expect(err).To(HaveOccurred(),
				"insecure cipher suite %s (0x%04x) must not be supported (CWE-327)", suite.Name, suite.ID)
		}
	})

	It("should accept only secure cipher suites", func() {
		// Verify that all entries in supportedCipherSuites come from tls.CipherSuites()
		secureSuites := tls.CipherSuites()
		for _, secure := range secureSuites {
			id, err := CipherSuiteStringToID(secure.Name)
			Expect(err).ToNot(HaveOccurred(),
				"secure cipher suite %s must be supported", secure.Name)
			Expect(id).To(Equal(secure.ID))
		}
	})
})

var _ = Describe("#TLSVersionToConstant security tests (CWE-327)", func() {
	It("should reject deprecated TLS versions", func() {
		deprecatedVersions := []configv1.TLSProtocolVersion{
			configv1.VersionTLS10,
			configv1.VersionTLS11,
		}

		for _, version := range deprecatedVersions {
			// These should default to TLS 1.2, not actually support the deprecated version
			result, err := TLSVersionToConstant(version)
			Expect(err).ToNot(HaveOccurred())
			// Deprecated versions should not be returned; defaults to TLS 1.2
			Expect(result).To(Equal(uint16(tls.VersionTLS12)),
				"deprecated TLS version %s must not be supported; should default to TLS 1.2 (CWE-327)", version)
		}
	})

	It("should support only secure TLS versions", func() {
		secureVersions := []struct {
			version  configv1.TLSProtocolVersion
			expected uint16
		}{
			{configv1.VersionTLS12, uint16(tls.VersionTLS12)},
			{configv1.VersionTLS13, uint16(tls.VersionTLS13)},
		}

		for _, tc := range secureVersions {
			result, err := TLSVersionToConstant(tc.version)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(tc.expected),
				"secure TLS version %s must map to %v", tc.version, tc.expected)
		}
	})
})

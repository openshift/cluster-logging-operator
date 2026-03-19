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
	It("should return the profile ciphers when they are defined", func() {
		Expect(TLSCiphers(configv1.TLSProfileSpec{Ciphers: []string{"a", "b"}})).To(Equal([]string{"a", "b"}))
	})
})

var _ = Describe("#MinTLSVersion", func() {
	It("should return the default min TLS version when not defined", func() {
		Expect(string(DefaultMinTLSVersion)).To(Equal(MinTLSVersion(configv1.TLSProfileSpec{})))
	})
	It("should return the profile min TLS version when defined", func() {
		Expect(string(configv1.VersionTLS13)).To(Equal(MinTLSVersion(configv1.TLSProfileSpec{MinTLSVersion: configv1.VersionTLS13})))
	})
})

var _ = Describe("#TLSGroups", func() {
	It("should return the default groups when none are provided", func() {
		Expect(TLSGroups(nil)).To(Equal(DefaultTLSGroups))
	})
	It("should return the default groups when an empty slice is provided", func() {
		Expect(TLSGroups([]string{})).To(Equal(DefaultTLSGroups))
	})
	It("should return the provided groups when they are defined", func() {
		groups := []string{"X25519", "secp256r1"}
		Expect(TLSGroups(groups)).To(Equal([]string{"X25519", "secp256r1"}))
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

package tls_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	. "github.com/openshift/cluster-logging-operator/internal/tls"
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

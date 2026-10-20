package tls

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
)

// withFIPSFile points the FIPS detection file at a fixture containing
// contents for the duration of fn, restoring the original path afterwards.
func withFIPSFile(contents string, fn func()) {
	orig := fipsEnabledFile
	dir, err := os.MkdirTemp("", "fips")
	Expect(err).ToNot(HaveOccurred())
	file := filepath.Join(dir, "fips_enabled")
	Expect(os.WriteFile(file, []byte(contents), 0o644)).To(Succeed())
	fipsEnabledFile = file
	defer func() {
		fipsEnabledFile = orig
		Expect(os.RemoveAll(dir)).To(Succeed())
	}()
	fn()
}

var _ = Describe("#FIPSEnabled", func() {
	It("reports FIPS enabled when the kernel flag is 1", func() {
		withFIPSFile("1\n", func() {
			Expect(FIPSEnabled()).To(BeTrue())
		})
	})

	It("reports FIPS disabled when the kernel flag is 0", func() {
		withFIPSFile("0\n", func() {
			Expect(FIPSEnabled()).To(BeFalse())
		})
	})

	It("reports FIPS disabled when the kernel flag cannot be read", func() {
		orig := fipsEnabledFile
		fipsEnabledFile = filepath.Join(os.TempDir(), "fips-enabled-does-not-exist")
		defer func() { fipsEnabledFile = orig }()
		Expect(FIPSEnabled()).To(BeFalse())
	})
})

var _ = Describe("#FilterGroupsForFIPS", func() {
	It("drops the post-quantum hybrid groups when FIPS is enabled", func() {
		groups := []configv1.TLSGroup{
			configv1.TLSGroupX25519MLKEM768,
			configv1.TLSGroupX25519,
			configv1.TLSGroupSecP256r1MLKEM768,
			configv1.TLSGroupSecP256r1,
			configv1.TLSGroupSecP384r1MLKEM1024,
			configv1.TLSGroupSecP384r1,
		}
		withFIPSFile("1\n", func() {
			Expect(FilterGroupsForFIPS(groups)).To(Equal([]configv1.TLSGroup{
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
				configv1.TLSGroupSecP384r1,
			}))
		})
	})

	It("keeps groups unchanged when FIPS is disabled", func() {
		groups := []configv1.TLSGroup{configv1.TLSGroupX25519MLKEM768, configv1.TLSGroupX25519}
		withFIPSFile("0\n", func() {
			Expect(FilterGroupsForFIPS(groups)).To(Equal(groups))
		})
	})

	It("returns an empty list when no group is FIPS-approved", func() {
		groups := []configv1.TLSGroup{configv1.TLSGroupX25519MLKEM768, configv1.TLSGroupSecP256r1MLKEM768}
		withFIPSFile("1\n", func() {
			Expect(FilterGroupsForFIPS(groups)).To(BeEmpty())
		})
	})
})

var _ = Describe("#TLSGroupsToOpenSSL on a FIPS-enabled host", func() {
	It("omits the post-quantum hybrid groups from the generated list", func() {
		withFIPSFile("1\n", func() {
			groups := []configv1.TLSGroup{
				configv1.TLSGroupX25519MLKEM768,
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
				configv1.TLSGroupSecP384r1,
			}
			Expect(TLSGroupsToOpenSSL(groups)).To(Equal("X25519:prime256v1:secp384r1"))
		})
	})

	It("leaves the generated list unchanged when FIPS is disabled", func() {
		withFIPSFile("0\n", func() {
			groups := []configv1.TLSGroup{configv1.TLSGroupX25519MLKEM768, configv1.TLSGroupX25519}
			Expect(TLSGroupsToOpenSSL(groups)).To(Equal("X25519MLKEM768:X25519"))
		})
	})

	It("omits the hybrid group from the default intermediate profile", func() {
		withFIPSFile("1\n", func() {
			openSSLGroups := TLSGroupsToOpenSSL(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].Groups)
			Expect(openSSLGroups).ToNot(ContainSubstring("MLKEM"))
			Expect(openSSLGroups).To(ContainSubstring("X25519"))
			Expect(openSSLGroups).To(ContainSubstring("prime256v1"))
			Expect(openSSLGroups).To(ContainSubstring("secp384r1"))
		})
	})
})

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMainSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Certificate Suite")
}

var _ = Describe("waitForMetricsCertificate", func() {
	var (
		dir      string
		certPath string
		keyPath  string

		interval = 10 * time.Millisecond
		timeout  = 250 * time.Millisecond
	)

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		certPath = filepath.Join(dir, "tls.crt")
		keyPath = filepath.Join(dir, "tls.key")
	})

	It("returns immediately when the certificate and key exist", func() {
		Expect(os.WriteFile(certPath, []byte("cert"), 0o600)).To(Succeed())
		Expect(os.WriteFile(keyPath, []byte("key"), 0o600)).To(Succeed())

		Expect(waitForMetricsCertificate(certPath, keyPath, interval, timeout)).To(Succeed())
	})

	It("waits until the certificate and key appear", func() {
		Expect(os.WriteFile(certPath, []byte("cert"), 0o600)).To(Succeed())

		go func() {
			time.Sleep(50 * time.Millisecond)
			Expect(os.WriteFile(keyPath, []byte("key"), 0o600)).To(Succeed())
		}()

		Expect(waitForMetricsCertificate(certPath, keyPath, interval, timeout)).To(Succeed())
	})

	It("waits until the files are non-empty", func() {
		Expect(os.WriteFile(certPath, nil, 0o600)).To(Succeed())
		Expect(os.WriteFile(keyPath, []byte("key"), 0o600)).To(Succeed())

		go func() {
			time.Sleep(50 * time.Millisecond)
			Expect(os.WriteFile(certPath, []byte("cert"), 0o600)).To(Succeed())
		}()

		Expect(waitForMetricsCertificate(certPath, keyPath, interval, timeout)).To(Succeed())
	})

	It("times out when the files never appear", func() {
		err := waitForMetricsCertificate(certPath, keyPath, interval, timeout)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("metrics certificate files not available"))
	})
})

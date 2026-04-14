package tls

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTLS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLS Suite")
}

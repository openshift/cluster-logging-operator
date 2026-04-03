package tlsprofile_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTLSProfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][controller][tlsprofile] Suite")
}

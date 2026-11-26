package namespace

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNamespaceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "must-gather/internal/namespace suite")
}

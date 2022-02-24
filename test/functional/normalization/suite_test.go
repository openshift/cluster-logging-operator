package normalization

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNormalization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterLogging Functional Normalization Suite")
}

package multiple

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Outputs][Multiple] Test Suite")
}

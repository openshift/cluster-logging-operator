package input_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInputs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inputs Suite")
}

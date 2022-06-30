package elements

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation Elements")
}

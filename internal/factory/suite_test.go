package factory

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resource Factory")
}

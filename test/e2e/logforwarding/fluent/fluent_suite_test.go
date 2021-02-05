package fluent_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluent Suite")
}

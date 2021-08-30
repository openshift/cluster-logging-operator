package output_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBufferConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buffer Conf Generation")
}

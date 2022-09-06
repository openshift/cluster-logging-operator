package pipelines

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPipelinesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Pipelines] Suite")
}

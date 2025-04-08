package pipelines

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPipelinesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Pipelines] Suite")
}

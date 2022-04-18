package elasticsearch

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFunctionalElasticsearchOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][outputs][elasticsearch] Suite")
}

package prune

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPruneFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[filter][prune] Unit Tests")
}

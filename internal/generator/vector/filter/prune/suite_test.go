package prune

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPruneFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[filter][prune] Unit Tests")
}

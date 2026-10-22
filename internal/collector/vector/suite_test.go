package vector_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][collector][vector] suite")
}

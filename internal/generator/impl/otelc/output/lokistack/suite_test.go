package lokistack_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLokiStack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LokiStack Output Suite")
}

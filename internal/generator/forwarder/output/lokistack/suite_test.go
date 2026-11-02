package lokistack_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLokiStackCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LokiStack Common Suite")
}

package observability_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][api][observability] suite")
}

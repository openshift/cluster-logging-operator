package indexmanagement_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIndexmanagement(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Indexmanagement Suite")
}

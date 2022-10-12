package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestV1Logging(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[apis][logging][v1] Suite")
}

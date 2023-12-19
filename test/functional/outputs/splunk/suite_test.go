package splunk

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputSplunk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Outputs][Splunk] Suite")
}

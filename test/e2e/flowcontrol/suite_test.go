package flowcontrol

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[FlowControl][E2E] Suite")
}

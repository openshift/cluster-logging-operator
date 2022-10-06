package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestInternalValidationsClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][validations][clustrelogforwarder] Suite")
}

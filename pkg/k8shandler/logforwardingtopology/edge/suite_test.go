package edge_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEdge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Edge Logforwarding Topology Suite")
}

package clusterinfo

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterInfo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][clusterinfo] Suite")
}

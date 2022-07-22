package consoleplugin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConsolePlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterLogging E2E Suite - Console Plugin")
}

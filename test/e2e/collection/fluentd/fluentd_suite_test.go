package fluentd

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluentd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterLogging E2E Suite - Collection Fluentd")
}

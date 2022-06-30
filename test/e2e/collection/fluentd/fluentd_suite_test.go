package fluentd

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

func TestFluentd(t *testing.T) {
	RegisterFailHandler(Fail)
	config.DefaultReporterConfig.SlowSpecThreshold = 120
	RunSpecs(t, "ClusterLogging E2E Suite - Collection Fluentd")
}

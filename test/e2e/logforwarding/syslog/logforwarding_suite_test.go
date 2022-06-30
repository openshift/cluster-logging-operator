package syslog

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	config.DefaultReporterConfig.SlowSpecThreshold = 120
	RunSpecs(t, "ClusterLogForwarder E2E Suite - Syslog")
}

package sysloglegacy

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogForwarder E2E Suite - Syslog Legacy"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-syslog-legacy.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}

package kafka

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestLogForwarding(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogForwarder E2E Suite - Kafka"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-kafka.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}

package fluentd

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestFluentd(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogging E2E Suite - Collection Fluentd"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-collection-fluentd.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}

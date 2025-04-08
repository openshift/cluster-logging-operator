package syslog

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputs(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "[Functional][Outputs][Syslog] Test Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-functional-outputs-syslog.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}

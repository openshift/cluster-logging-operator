package collection

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestLogForwarding(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "[Functional][Collection] Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-functional-collection.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}

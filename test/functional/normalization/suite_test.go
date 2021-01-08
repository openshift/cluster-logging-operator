package normalization

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestNormalization(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogging Functional Normalization Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-normalization.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}

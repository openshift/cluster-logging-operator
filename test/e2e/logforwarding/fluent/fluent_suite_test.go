package fluent_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

func TestFluent(t *testing.T) {
	RegisterFailHandler(Fail)
	config.DefaultReporterConfig.SlowSpecThreshold = 120
	RunSpecs(t, "Fluent Suite")
}

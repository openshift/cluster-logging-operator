package cloudwatch

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputCloudwatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Outputs][Cloudwatch] Suite")
}

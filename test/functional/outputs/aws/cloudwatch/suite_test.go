package cloudwatch

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputCloudwatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Outputs][AWS][Cloudwatch] Suite")
}

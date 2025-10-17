package s3

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Functional][Outputs][AWS][S3] Suite")
}

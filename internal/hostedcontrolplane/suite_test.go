package hostedcontrolplane

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestV1Logging(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][clusterinfo] Suite")
}

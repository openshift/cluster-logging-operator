package lokistack

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLokistack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[e2e][logforwarding][lokistack] Suite")
}

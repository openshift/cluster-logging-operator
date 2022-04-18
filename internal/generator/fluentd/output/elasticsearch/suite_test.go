package elasticsearch

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluentdConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][generator][fluentd][output][elasticsearch] Suite")
}

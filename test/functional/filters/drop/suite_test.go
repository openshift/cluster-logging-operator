package drop

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFiltersDrop(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][filters][drop]")
}

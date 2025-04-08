package apiaudit

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFiltersKubeAPIAudit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][filters][apiaudit]")
}

package dashboard_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDashboards(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][metrics][dashboard] suite")
}

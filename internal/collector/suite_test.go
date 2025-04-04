package collector_test

import (
	"embed"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	//go:embed cw_multiple_credentials cw_single_credential
	credFiles embed.FS
)

func TestCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][collector] suite")
}

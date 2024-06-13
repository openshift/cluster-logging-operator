package common

import (
	"embed"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	//go:embed *.toml *.sh
	embedFS embed.FS
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][generator][vector][common] Suite")
}

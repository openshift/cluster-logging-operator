package apivalidations

import (
	"embed"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	//go:embed *.yaml
	tomlContent embed.FS
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[e2e][collection][apivalidations] Suite")
}

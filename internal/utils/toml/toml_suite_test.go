package toml

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestToml(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils TOML Suite")
}

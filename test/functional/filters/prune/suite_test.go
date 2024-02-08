package prune

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFiltersPrune(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][filters][prune]")
}

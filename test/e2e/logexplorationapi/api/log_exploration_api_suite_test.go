package api

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLogExplorationApi(t *testing.T) {

	RegisterFailHandler(Fail)

	RunSpecs(t, "LogExplorationApi Suite")
}

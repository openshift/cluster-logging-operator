package kafka

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFunctionalKafkaOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][outputs][kafka] Suite")
}

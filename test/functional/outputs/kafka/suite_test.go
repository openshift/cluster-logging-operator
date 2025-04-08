package kafka

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFunctionalKafkaOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][outputs][kafka] Suite")
}

package logstash

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFunctionalLogStashOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][outputs][logstash] Suite")
}

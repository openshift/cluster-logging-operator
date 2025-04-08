package logstash

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFunctionalLogStashOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][outputs][logstash] Suite")
}

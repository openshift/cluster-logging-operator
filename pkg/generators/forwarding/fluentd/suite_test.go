package fluentd_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluentdGenerators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluentd Generator Suite")
}

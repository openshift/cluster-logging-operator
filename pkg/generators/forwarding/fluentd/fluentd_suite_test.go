package fluentd_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluentd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluentd Suite")
}

package source_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSourceConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Unit][Internal][Generator][Fluentd][Source] Suite")
}

package s3_test

import (
	"embed"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	//go:embed files/*
	testFiles embed.FS
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][generator][vector][output][s3] Suite")
}

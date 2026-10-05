package syslog

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSyslog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Syslog Log Forwarding E2E Suite")
}

package aws_test

import (
	"embed"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	//go:embed cw_multiple_credentials cw_single_credentials cw_assume_role_single
	credFiles embed.FS
)

func TestCollectorAWS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][collector][aws] suite")
}

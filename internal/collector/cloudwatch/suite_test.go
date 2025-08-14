package cloudwatch_test

import (
	"embed"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	//go:embed cw_multiple_credentials cw_single_credential cw_assume_role_single
	credFiles embed.FS
)

func TestCollectorCloudwatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][collector][cloudwatch] suite")
}

package elasticsearch

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCLOElasticsearchLogStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite - CLO Managed Elasticsearch Log Store")
}

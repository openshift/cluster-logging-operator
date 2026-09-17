package version_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/version"
	projectVersion "github.com/openshift/cluster-logging-operator/version"
)

var _ = Describe("VersionCollector", func() {
	var (
		tmpDir  string
		destDir api.Path
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "version-collector-test")
		Expect(err).ToNot(HaveOccurred())
		destDir = api.NewPath(tmpDir)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	It("should write a version file with the product name and version", func() {
		collector := version.NewCollector(api.NewLogger(GinkgoWriter), destDir)
		err := collector.Collect(context.Background())
		Expect(err).ToNot(HaveOccurred())

		content, err := os.ReadFile(filepath.Join(tmpDir, "version"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal("Red Hat OpenShift Logging/must-gather\n" + projectVersion.Version + "\n"))
	})

	It("should report its name", func() {
		collector := version.NewCollector(api.NewLogger(GinkgoWriter), destDir)
		Expect(collector.Name()).To(Equal("VersionCollector"))
	})
})

package version

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/version"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const productName = "Red Hat OpenShift Logging"

type Collector struct {
	logger  api.Logger
	destDir api.Path
}

func NewCollector(logger api.Logger, destDir api.Path) *Collector {
	return &Collector{
		logger:  logger,
		destDir: destDir,
	}
}

func (c *Collector) Name() string {
	return "VersionCollector"
}

func (c *Collector) Collect(_ context.Context, _ ...schema.GroupVersionResource) error {
	defer c.logger.Begin("writing version file ...")()

	content := fmt.Sprintf("%s/must-gather\n%s\n", productName, version.Version)
	versionFile := c.destDir.Add("version")
	return versionFile.WriteFile([]byte(content))
}

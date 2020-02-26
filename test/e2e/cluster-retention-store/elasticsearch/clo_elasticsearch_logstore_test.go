package elasticsearch

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("CLO Elasticsearch Log Store", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		e2e = helpers.NewE2ETestFramework()
		log = "Log me errors one more time"
	)

	Describe("when ClusterLogging is configured with a collector and elasticsearch logstore", func() {
		BeforeEach(func() {
			components := []helpers.LogComponentType{
				helpers.ComponentTypeCollector,
				helpers.ComponentTypeStore,
			}

			cr := helpers.NewClusterLogging(components...)
			if err := e2e.SetupClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

			if err := e2e.DeployLogGeneratorFor(log); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}
		})

		AfterEach(func() {
			e2e.Cleanup()
		}, helpers.DefaultCleanUpTimeout)

		It("should default to collecting logs to the spec'd logstore as retreived on the generator", func() {
			Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
			Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
			// Expect(e2e.LogStore.HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			Expect(e2e.LogStore.HasAppLogEntry(log, helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find specific application log entry")
		})
	})
})

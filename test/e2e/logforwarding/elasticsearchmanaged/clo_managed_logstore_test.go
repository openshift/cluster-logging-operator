package elasticsearchmanaged

import (
	"fmt"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("CLO Managed ClusterLogForwarder", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		e2e = helpers.NewE2ETestFramework()
	)

	Describe("when ClusterLogging is configured with a collector, LogStore, and no explicit 'forwarder'", func() {

		BeforeEach(func() {
			if err := e2e.DeployLogGenerator(); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}

			components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
			if err := e2e.SetupClusterLogging(components...); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion([]string{"fluentd", "elasticsearch"})
		}, helpers.DefaultCleanUpTimeout)

		It("should default to forwarding logs to the spec'd logstore", func() {
			Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
			Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")

			//verify infra namespaces are not stored to their own index
			elasticSearch := helpers.ElasticLogStore{Framework: e2e}
			if indices, err := elasticSearch.Indices(); err != nil {
				Fail(fmt.Sprintf("Error fetching indices: %v", err))
			} else {
				for _, index := range indices {
					if strings.HasPrefix(index.Name, "project.openshift") || strings.HasPrefix(index.Name, "project.kube") {
						Fail(fmt.Sprintf("Found an infra namespace that was not stored in an infra index: %s", index.Name))
					}
				}
			}

		})

	})

})

package fluentd

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"runtime"

	. "github.com/onsi/ginkgo"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[Collection] Provides only a fluentd daemonset", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e        = helpers.NewE2ETestFramework()
		components []helpers.LogComponentType
	)

	Describe("when ClusterLogging is configured only with a collection spec", func() {

		BeforeEach(func() {
			if err := e2e.DeployLogGenerator(); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}

			components = []helpers.LogComponentType{helpers.ComponentTypeCollector}
			if err := e2e.SetupClusterLogging(components...); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}

		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{constants.CollectorName})
		}, helpers.DefaultCleanUpTimeout)

		It("should default to a running collector", func() {
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}
		})
	})
})

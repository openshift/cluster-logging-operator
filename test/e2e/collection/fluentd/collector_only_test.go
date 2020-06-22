package fluentd

import (
  "fmt"
  "runtime"

  . "github.com/onsi/ginkgo"

  "github.com/openshift/cluster-logging-operator/pkg/logger"
  "github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("Fluentd Only Deployment", func() {
  _, filename, _, _ := runtime.Caller(0)
  logger.Infof("Running %s", filename)
  var (
    e2e        = helpers.NewE2ETestFramework()
    components []helpers.LogComponentType
  )

  Describe("when ClusterLogging is configured only with a collector", func() {

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
      e2e.WaitForCleanupCompletion([]string{"fluentd"})
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

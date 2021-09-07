package api

import (
	"fmt"
	apps "k8s.io/api/apps/v1"
	"runtime"

	"github.com/ViaQ/logerr/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("Deploys API and collect logs from elasticsearch", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e                         = helpers.NewE2ETestFramework()
		components                  []helpers.LogComponentType
		err                         error
		logExplorationAPiDeployment *apps.Deployment
	)
	Describe("Deploying API and collecting logs from elasticsearch", func() {
		BeforeEach(func() {

			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector, helpers.ComponentTypeStore, helpers.ComponentLogAPI)
			if err = e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				return
			}
			if err = e2e.WaitFor(helpers.ComponentTypeCollector); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
				return
			}
			if err = e2e.WaitFor(helpers.ComponentTypeStore); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeStore, err))
				return
			}
			if e2e.ClusterLogging.ObjectMeta.Annotations["api-enabled"] == "true" {
				logExplorationAPiDeployment, err = e2e.DeployLogExplorationAPI()
				if err != nil {
					Fail(fmt.Sprintf("Failed to deploy the compoenent %s: %v", helpers.ComponentLogAPI, err))
				}
			}
			if err = e2e.WaitFor(helpers.ComponentLogAPI); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentLogAPI, err))
				return
			}
		})

		It("should collect logs from the endpoint", func() {
			for _, component := range components {
				if err = e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

			Expect(e2e.LogStores[logExplorationAPiDeployment.GetName()].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs with ")
		})

		AfterEach(func() {
			e2e.Cleanup()
		}, helpers.DefaultCleanUpTimeout)
	})
})

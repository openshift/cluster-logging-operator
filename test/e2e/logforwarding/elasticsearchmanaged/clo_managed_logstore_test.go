package elasticsearchmanaged

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.NewLogger("e2e-logforwarding").Info("Running ", "filename", filename)
	var (
		e2e = framework.NewE2ETestFramework()
	)

	Describe("when the output is a CLO managed elasticsearch and no explicit forwarder is configured", func() {

		DeployLoggingWithComponents := func(components []helpers.LogComponentType) {
			if err := e2e.SetupClusterLogging(components...); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}
		}

		BeforeEach(func() {
			ns := e2e.CreateTestNamespace()
			appLabels := map[string]string{
				"myapp":                    "test-log-generator",
				"myapp.kubernetes.io/name": "test-log-generator",
			}
			if err := e2e.DeployLogGeneratorWithNamespaceAndLabels(ns, appLabels); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}
		})

		AssertBehaviours := func() {
			It("should default to forwarding logs to the spec'd logstore", func() {
				Expect(e2e.LogStores["elasticsearch"].HasInfraStructureLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStores["elasticsearch"].HasApplicationLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")

				//verify infra namespaces are not stored to their own index
				elasticSearch := framework.ElasticLogStore{Framework: e2e}
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
		}

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(constants.OpenshiftNS, []string{constants.CollectorName, "elasticsearch"})
		}, framework.DefaultCleanUpTimeout)

		Describe("for fluentd collector", func() {
			BeforeEach(func() {
				DeployLoggingWithComponents([]helpers.LogComponentType{helpers.ComponentTypeCollectorFluentd, helpers.ComponentTypeStore})
			})
			AssertBehaviours()
		})
		Describe("for vector collector", func() {
			BeforeEach(func() {
				DeployLoggingWithComponents([]helpers.LogComponentType{helpers.ComponentTypeCollectorVector, helpers.ComponentTypeStore})
			})
			AssertBehaviours()
		})
	})

})

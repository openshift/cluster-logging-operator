package elasticsearchmanaged

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"runtime"
	"strings"

	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e = framework.NewE2ETestFramework()
	)

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

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(constants.WatchNamespace, []string{constants.CollectorName, "elasticsearch"})
	}, framework.DefaultCleanUpTimeout)

	DescribeTable("when the output is a CLO managed elasticsearch and no explicit forwarder is configured should default to forwarding logs to the spec'd logstore", func(collectorType helpers.LogComponentType) {
		components := []helpers.LogComponentType{collectorType, helpers.ComponentTypeStore}
		if err := e2e.SetupClusterLogging(components...); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
		}
		for _, component := range components {
			if err := e2e.WaitFor(component); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
			}
		}

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
	},
		Entry("using fluentd collector", helpers.ComponentTypeCollectorFluentd),
		Entry("using vector collector", helpers.ComponentTypeCollectorVector),
	)

})

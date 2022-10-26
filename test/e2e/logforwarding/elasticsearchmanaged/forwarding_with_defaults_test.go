package elasticsearchmanaged

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e         = framework.NewE2ETestFramework()
		generatorNS string
	)

	Describe("ClusterLogging with default store", func() {
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
		SetupLogGeneratorWithLabels := func() {
			appLabels := map[string]string{
				"myapp":                    "test-log-generator",
				"myapp.kubernetes.io/name": "test-log-generator",
				"logFormat":                "redhat",
			}
			var err error
			generatorNS, _, err = e2e.DeployJsonLogGenerator(map[string]string{
				"level":   "debug",
				"logtext": "hey, this is a log line",
			}, appLabels)
			if err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}
		}

		AssertLogStoreHasIndex := func(store framework.LogStore, prefix string) {
			estore := store.(*framework.ElasticLogStore)
			var indices framework.Indices
			pollErr := wait.PollImmediate(5*time.Second, 10*time.Minute, func() (found bool, err error) {
				indices, err = estore.Indices()
				if err != nil {
					log.Error(err, "Error retrieving indices from elasticsearch")
					return false, nil
				}
				log.V(2).Info("indices", "indices", indices)
				for _, index := range indices {
					if strings.HasPrefix(index.Name, prefix) {
						found = true
					}
				}
				return found, nil
			})
			Expect(pollErr).To(BeNil(), fmt.Sprintf("Unable to find logs in store, indices: %v", indices))
		}

		BeforeEach(func() {
			SetupLogGeneratorWithLabels()
		})
		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(constants.WatchNamespace, []string{constants.CollectorName, "elasticsearch"})
			e2e.WaitForCleanupCompletion(generatorNS, []string{"component", "test"})
		}, framework.DefaultCleanUpTimeout)

		Context("forwarding logs to default output", func() {
			Context("with StructuredTypeKey set in outputDefaults", func() {
				DescribeTable("should send logs to index set in labels", func(collectorType helpers.LogComponentType) {
					DeployLoggingWithComponents([]helpers.LogComponentType{collectorType, helpers.ComponentTypeStore})
					forwarder := testruntime.NewClusterLogForwarder()
					forwarder.Spec = logging.ClusterLogForwarderSpec{
						Inputs: []logging.InputSpec{
							{
								Name: "my-application",
								Application: &logging.Application{
									Namespaces: []string{generatorNS},
								},
							},
						},
						OutputDefaults: &logging.OutputDefaults{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeKey: "kubernetes.labels.logFormat",
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{logging.OutputNameDefault},
								InputRefs:  []string{"my-application"},
								Parse:      "json",
							},
						},
					}
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeStore, collectorType}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}

					store := e2e.LogStores["elasticsearch"]
					AssertLogStoreHasIndex(store, "app-redhat")
				},
					Entry("with fluentd", helpers.ComponentTypeCollectorFluentd),
					Entry("with vector", helpers.ComponentTypeCollectorVector),
				)

			})
			Context("with StructuredTypeName set in outputDefaults", func() {
				DescribeTable("should send logs to index set in StructuredTypeName", func(collectorType helpers.LogComponentType) {
					DeployLoggingWithComponents([]helpers.LogComponentType{collectorType, helpers.ComponentTypeStore})
					IndexName := "testindex"
					forwarder := testruntime.NewClusterLogForwarder()
					forwarder.Spec = logging.ClusterLogForwarderSpec{
						Inputs: []logging.InputSpec{
							{
								Name: "my-service",
								Application: &logging.Application{
									Namespaces: []string{generatorNS},
								},
							},
						},
						OutputDefaults: &logging.OutputDefaults{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeName: IndexName,
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{logging.OutputNameDefault},
								InputRefs:  []string{"my-service"},
								Parse:      "json",
							},
						},
					}
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
					}
					components := []helpers.LogComponentType{collectorType, helpers.ComponentTypeStore}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}

					store := e2e.LogStores["elasticsearch"]
					AssertLogStoreHasIndex(store, "app-testindex")
				},
					Entry("with fluentd", helpers.ComponentTypeCollectorFluentd),
					Entry("with vector", helpers.ComponentTypeCollectorVector),
				)
			})
		})
	})
})

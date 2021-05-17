package elasticsearchmanaged

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e = helpers.NewE2ETestFramework()
	)

	Describe("ClusterLogging with default store", func() {
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
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluentd", "elasticsearch"})
		}, helpers.DefaultCleanUpTimeout)

		Context("forwarding logs to default output", func() {
			Context("with IndexKey set in outputDefaults", func() {
				BeforeEach(func() {
					forwarder := &logging.ClusterLogForwarder{
						TypeMeta: metav1.TypeMeta{
							Kind:       logging.ClusterLogForwarderKind,
							APIVersion: logging.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "instance",
						},
						Spec: logging.ClusterLogForwarderSpec{
							OutputDefaults: &logging.OutputDefaults{
								Elasticsearch: &logging.Elasticsearch{
									StructuredIndexKey: "kubernetes.labels.component",
								},
							},
							Pipelines: []logging.PipelineSpec{
								{
									Name:       "test-app",
									OutputRefs: []string{logging.OutputNameDefault},
									InputRefs:  []string{logging.InputNameApplication},
								},
							},
						},
					}
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
				})
				It("should send logs to index set in labsls ", func() {
					store := e2e.LogStores["elasticsearch"]
					estore := store.(*helpers.ElasticLogStore)
					var indices helpers.Indices
					var err error
					err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
						indices, err = estore.Indices()
						if err != nil {
							log.Error(err, "Error retrieving indices from elasticsearch")
							return false, nil
						}
						found := false
						for _, index := range indices {
							if index.Name == "test" {
								found = true
							}
						}
						return found, nil
					})
					Expect(err).To(BeNil())
				})
			})
			Context("with IndexName set in outputDefaults", func() {
				IndexName := "testindex"
				BeforeEach(func() {
					forwarder := &logging.ClusterLogForwarder{
						TypeMeta: metav1.TypeMeta{
							Kind:       logging.ClusterLogForwarderKind,
							APIVersion: logging.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "instance",
						},
						Spec: logging.ClusterLogForwarderSpec{
							OutputDefaults: &logging.OutputDefaults{
								Elasticsearch: &logging.Elasticsearch{
									StructuredIndexName: IndexName,
								},
							},
							Pipelines: []logging.PipelineSpec{
								{
									Name:       "test-app",
									OutputRefs: []string{logging.OutputNameDefault},
									InputRefs:  []string{logging.InputNameApplication},
								},
							},
						},
					}
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
				})
				It("should send logs to index set in IndexName ", func() {
					store := e2e.LogStores["elasticsearch"]
					estore := store.(*helpers.ElasticLogStore)
					var indices helpers.Indices
					var err error
					err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
						indices, err = estore.Indices()
						if err != nil {
							log.Error(err, "Error retrieving indices from elasticsearch")
							return false, nil
						}
						found := false
						for _, index := range indices {
							if index.Name == IndexName {
								found = true
							}
						}
						return found, nil
					})
					Expect(err).To(BeNil())
				})
			})
		})

	})
})

package elasticsearchmanaged

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e = framework.NewE2ETestFramework()
	)

	Describe("ClusterLogging with default store", func() {
		BeforeEach(func() {
			appLabels := map[string]string{"logFormat": "redhat"}
			ns, pod, err := e2e.DeployJsonLogGenerator(map[string]string{
				"level":   "debug",
				"logtext": "hey, this is a log line",
			})
			if err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}
			for k, v := range appLabels {
				if _, err := oc.Literal().From("oc -n %s label pods %s %s=%s --overwrite=true", ns, pod, k, v).Run(); err != nil {
					Fail(fmt.Sprintf("Timed out waiting for the log generator to apply labels: %v", err))
				}
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
			e2e.WaitForCleanupCompletion(constants.OpenshiftNS, []string{"fluentd", "elasticsearch"})
		}, framework.DefaultCleanUpTimeout)

		Context("forwarding logs to default output", func() {
			Context("with TypeKey set in outputDefaults", func() {
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
									StructuredTypeKey: "kubernetes.labels.logFormat",
								},
							},
							Pipelines: []logging.PipelineSpec{
								{
									Name:       "test-app",
									OutputRefs: []string{logging.OutputNameDefault},
									InputRefs:  []string{logging.InputNameApplication},
									Parse:      "json",
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
					estore := store.(*framework.ElasticLogStore)
					var indices framework.Indices
					var err error
					err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
						indices, err = estore.Indices()
						if err != nil {
							log.Error(err, "Error retrieving indices from elasticsearch")
							return false, nil
						}
						found := false
						log.V(2).Info("indices", "indices", indices)
						for _, index := range indices {
							if strings.HasPrefix(index.Name, "app-redhat") {
								found = true
							}
						}
						return found, nil
					})
					log.V(2).Info("error", "error", err)
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
									StructuredTypeName: IndexName,
								},
							},
							Pipelines: []logging.PipelineSpec{
								{
									Name:       "test-app",
									OutputRefs: []string{logging.OutputNameDefault},
									InputRefs:  []string{logging.InputNameApplication},
									Parse:      "json",
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
					estore := store.(*framework.ElasticLogStore)
					var indices framework.Indices
					var err error
					err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
						indices, err = estore.Indices()
						if err != nil {
							log.Error(err, "Error retrieving indices from elasticsearch")
							return false, nil
						}
						found := false
						for _, index := range indices {
							if strings.HasPrefix(index.Name, "app-testindex") {
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

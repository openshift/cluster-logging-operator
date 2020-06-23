package elasticsearchunmanaged

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

var _ = Describe("ClusterLogForwarder", func() {

	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err error
		e2e = helpers.NewE2ETestFramework()
	)
	Describe("when ClusterLogging is configured with 'forwarder' to an administrator managed Elasticsearch", func() {

		BeforeEach(func() {
			rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
			logger.Debugf("Repo rootdir: %s", rootDir)
			err = e2e.DeployLogGenerator()
			if err != nil {
				Fail(fmt.Sprintf("Unable to deploy log generator. E: %s", err.Error()))
			}
			var pipelineSecret *corev1.Secret
			var elasticsearch *elasticsearch.Elasticsearch
			if elasticsearch, pipelineSecret, err = e2e.DeployAnElasticsearchCluster(rootDir); err != nil {
				Fail(fmt.Sprintf("Unable to deploy an elastic instance: %v", err))
			}

			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			forwarder := &logging.ClusterLogForwarder{
				TypeMeta: metav1.TypeMeta{
					Kind:       logging.ClusterLogForwarderKind,
					APIVersion: logging.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "instance",
				},
				Spec: logging.ClusterLogForwarderSpec{
					Outputs: []logging.OutputSpec{
						{
							Name: elasticsearch.Name,
							Secret: &logging.OutputSecretSpec{
								Name: pipelineSecret.ObjectMeta.Name,
							},
							Type: logging.OutputTypeElasticsearch,
							URL:  fmt.Sprintf("%s.%s.svc:9200", elasticsearch.Name, elasticsearch.Namespace),
						},
					},
					Pipelines: []logging.PipelineSpec{
						{
							Name:       "test-app",
							OutputRefs: []string{elasticsearch.Name},
							InputRefs:  []string{logging.InputNameApplication},
						},
						{
							Name:       "test-infra",
							OutputRefs: []string{elasticsearch.Name},
							InputRefs:  []string{logging.InputNameInfrastructure},
						},
						{
							Name:       "test-audit",
							OutputRefs: []string{elasticsearch.Name},
							InputRefs:  []string{logging.InputNameAudit},
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

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion([]string{"fluentd", "elasticsearch"})
		})

		It("should send logs to the forward.Output logstore", func() {
			Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
			Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
			Expect(e2e.LogStore.HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
		})

	})

})

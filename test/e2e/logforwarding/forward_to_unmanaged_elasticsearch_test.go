package logforwarding

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("LogForwarding", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err error
		e2e = helpers.NewE2ETestFramework()
	)
	Describe("when ClusterLogging is configured with 'forwarding' to an administrator managed Elasticsearch", func() {

		BeforeEach(func() {
			rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "/")
			logger.Debugf("Repo rootdir: %s", rootDir)
			e2e.DeployLogGenerator()
			var pipelineSecret *corev1.Secret
			var elasticsearch *elasticsearch.Elasticsearch
			if elasticsearch, pipelineSecret, err = e2e.DeployAnElasticsearchCluster(rootDir); err != nil {
				Fail(fmt.Sprintf("Unable to deploy an elastic instance: %v", err))
			}

			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
			cr.Spec.Forwarding = &cl.ForwardingSpec{
				Outputs: []cl.OutputSpec{
					cl.OutputSpec{
						Name: elasticsearch.Name,
						Secret: &cl.OutputSecretSpec{
							Name: pipelineSecret.ObjectMeta.Name,
						},
						Type:     cl.OutputTypeElasticsearch,
						Endpoint: fmt.Sprintf("%s.%s.svc:9200", elasticsearch.Name, elasticsearch.Namespace),
					},
				},
				Pipelines: []cl.PipelineSpec{
					cl.PipelineSpec{
						Name:       "test-app",
						OutputRefs: []string{elasticsearch.Name},
						SourceType: cl.LogSourceTypeApp,
					},
					cl.PipelineSpec{
						Name:       "test-infra",
						OutputRefs: []string{elasticsearch.Name},
						SourceType: cl.LogSourceTypeInfra,
					},
				},
			}
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
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
		})

		It("should send logs to the forward.Output logstore", func() {
			Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
			Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
		})

	})

})

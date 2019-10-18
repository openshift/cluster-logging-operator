package logforwarding

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	apps "k8s.io/api/apps/v1"
)

var _ = Describe("LogForwarding", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err              error
		fluentDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
	)
	BeforeEach(func() {
		e2e.DeployLogGenerator()
	})
	Describe("when ClusterLogging is configured with 'forwarding' to an administrator managed fluentd", func() {

		Context("and the receiver is unsecured", func() {

			BeforeEach(func() {
				if fluentDeployment, _, err = e2e.DeployFluendReceiver(); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				cr.Spec.Forwarding = &cl.ForwardingSpec{
					Outputs: []cl.OutputSpec{
						cl.OutputSpec{
							Name:     fluentDeployment.ObjectMeta.Name,
							Type:     cl.OutputTypeForward,
							Endpoint: fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
						},
					},
					Pipelines: []cl.PipelineSpec{
						cl.PipelineSpec{
							Name:       "test-app",
							OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
							SourceType: cl.LogSourceTypeApp,
						},
						cl.PipelineSpec{
							Name:       "test-infra",
							OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
							SourceType: cl.LogSourceTypeInfra,
						},
					},
				}
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})

			It("should send logs to the forward.Output logstore", func() {
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStore.HasApplicationStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
		})

	})

})

package fluent

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ClusterLogForwarder", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err              error
		fluentDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		rootDir          string
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
	})
	Describe("when ClusterLogging is configured with 'forwarder' to an administrator managed fluentd", func() {

		Context("and the receiver is unsecured", func() {

			BeforeEach(func() {
				fluentDeployment, err := e2e.DeployFluentdReceiver(rootDir, false)
				Expect(err).To(Succeed(), "DeployFluentdReceiver")

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				err = e2e.CreateClusterLogging(cr)
				Expect(err).To(Succeed(), "CreateClusterLogging")
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
								Name: fluentDeployment.ObjectMeta.Name,
								Type: logging.OutputTypeFluentForward,
								URL:  fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								InputRefs:  []string{logging.InputNameApplication},
							},
							{
								Name:       "test-infra",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								InputRefs:  []string{logging.InputNameInfrastructure},
							},
							{
								Name:       "test-audit",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								InputRefs:  []string{logging.InputNameAudit},
							},
						},
					},
				}
				logger.Infof("FIXME creating %v", forwarder)
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})

			It("should send logs to the forward.Output logstore", func() {
				Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStore.HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})
		})

		Context("and the receiver is secured", func() {

			BeforeEach(func() {
				if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, true); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}
				//sanity check
				initialWaitForLogsTimeout, _ := time.ParseDuration("30s")
				if exist, _ := e2e.LogStore.HasInfraStructureLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}
				if exist, _ := e2e.LogStore.HasApplicationLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
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
								Name: fluentDeployment.ObjectMeta.Name,
								Type: logging.OutputTypeFluentForward,
								URL:  fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
								Secret: &logging.OutputSecretSpec{
									Name: fluentDeployment.ObjectMeta.Name,
								},
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								InputRefs:  []string{logging.InputNameApplication},
							},
							{
								Name:       "test-infra",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								InputRefs:  []string{logging.InputNameInfrastructure},
							},
							{
								Name:       "test-audit",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								InputRefs:  []string{logging.InputNameAudit},
							},
						},
					},
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
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
				Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStore.HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			// FIXME(alanconway)
			// e2e.WaitForCleanupCompletion([]string{"fluent-receiver", "fluentd"})
		})

	})

})

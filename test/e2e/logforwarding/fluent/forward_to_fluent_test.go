package fluent

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"

	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LogForwarding", func() {
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
	Describe("when ClusterLogging is configured with 'forwarding' to an administrator managed fluentd", func() {

		Context("and the receiver is unsecured", func() {

			BeforeEach(func() {
				if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, false); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				forwarding := &logforward.LogForwarding{
					TypeMeta: metav1.TypeMeta{
						Kind:       logforward.LogForwardingKind,
						APIVersion: logforward.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logforward.ForwardingSpec{
						Outputs: []logforward.OutputSpec{
							logforward.OutputSpec{
								Name:     fluentDeployment.ObjectMeta.Name,
								Type:     logforward.OutputTypeForward,
								Endpoint: fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
							},
						},
						Pipelines: []logforward.PipelineSpec{
							logforward.PipelineSpec{
								Name:       "test-app",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								SourceType: logforward.LogSourceTypeApp,
							},
							logforward.PipelineSpec{
								Name:       "test-infra",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								SourceType: logforward.LogSourceTypeInfra,
							},
							logforward.PipelineSpec{
								Name:       "test-audit",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								SourceType: logforward.LogSourceTypeAudit,
							},
						},
					},
				}
				if err := e2e.CreateLogForwarding(forwarding); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
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
				forwarding := &logforward.LogForwarding{
					TypeMeta: metav1.TypeMeta{
						Kind:       logforward.LogForwardingKind,
						APIVersion: logforward.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logforward.ForwardingSpec{
						Outputs: []logforward.OutputSpec{
							logforward.OutputSpec{
								Name:     fluentDeployment.ObjectMeta.Name,
								Type:     logforward.OutputTypeForward,
								Endpoint: fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
								Secret: &logforward.OutputSecretSpec{
									Name: fluentDeployment.ObjectMeta.Name,
								},
							},
						},
						Pipelines: []logforward.PipelineSpec{
							logforward.PipelineSpec{
								Name:       "test-app",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								SourceType: logforward.LogSourceTypeApp,
							},
							logforward.PipelineSpec{
								Name:       "test-infra",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								SourceType: logforward.LogSourceTypeInfra,
							},
							logforward.PipelineSpec{
								Name:       "test-audit",
								OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
								SourceType: logforward.LogSourceTypeAudit,
							},
						},
					},
				}
				if err := e2e.CreateLogForwarding(forwarding); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
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
		})

	})

})

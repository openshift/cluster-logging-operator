package logforwarding

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)

	var (
		e2e              = helpers.NewE2ETestFramework()
		fluentDeployment *appsv1.Deployment
		err              error
		rootDir          string
	)

	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}

		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
	})

	Describe("after converting existing LogForwarding CR to ClusterLogForwader CR", func() {

		Context("and the receiver is unsecured", func() {

			BeforeEach(func() {
				fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, false)
				Expect(err).To(Succeed(), "DeployFluentdReceiver")

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				err = e2e.CreateClusterLogging(cr)
				Expect(err).To(Succeed(), "CreateClusterLogging")

				fw := &loggingv1alpha1.LogForwarding{
					TypeMeta: metav1.TypeMeta{
						Kind:       loggingv1alpha1.LogForwardingKind,
						APIVersion: loggingv1alpha1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: loggingv1alpha1.ForwardingSpec{
						Outputs: []loggingv1alpha1.OutputSpec{
							{
								Name:     fluentDeployment.GetName(),
								Type:     loggingv1alpha1.OutputTypeForward,
								Endpoint: fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
							},
						},
						Pipelines: []loggingv1alpha1.PipelineSpec{
							{
								Name:       "test-app",
								SourceType: loggingv1alpha1.LogSourceTypeApp,
								OutputRefs: []string{fluentDeployment.Name},
							},
							{
								Name:       "test-audit",
								SourceType: loggingv1alpha1.LogSourceTypeAudit,
								OutputRefs: []string{fluentDeployment.Name},
							},
							{
								Name:       "test-Infra",
								SourceType: loggingv1alpha1.LogSourceTypeInfra,
								OutputRefs: []string{fluentDeployment.Name},
							},
						},
					},
				}

				if err := e2e.CreateLogForwardingTechPreview(fw); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
				}

				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}
			})

			It("should send logs to the output", func() {
				name := fluentDeployment.GetName()
				Expect(e2e.LogStores[name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStores[name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStores[name].HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})
		})

		Context("and the receiver is secured", func() {

			BeforeEach(func() {
				fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, true)
				if err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}
				//sanity check
				initialWaitForLogsTimeout, _ := time.ParseDuration("30s")
				name := fluentDeployment.GetName()
				if exist, _ := e2e.LogStores[name].HasInfraStructureLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}
				if exist, _ := e2e.LogStores[name].HasApplicationLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}

				fw := &loggingv1alpha1.LogForwarding{
					TypeMeta: metav1.TypeMeta{
						Kind:       loggingv1alpha1.LogForwardingKind,
						APIVersion: loggingv1alpha1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: loggingv1alpha1.ForwardingSpec{
						Outputs: []loggingv1alpha1.OutputSpec{
							{
								Name:     fluentDeployment.GetName(),
								Type:     loggingv1alpha1.OutputTypeForward,
								Endpoint: fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
								Secret: &loggingv1alpha1.OutputSecretSpec{
									Name: fluentDeployment.ObjectMeta.Name,
								},
							},
						},
						Pipelines: []loggingv1alpha1.PipelineSpec{
							{
								Name:       "test-app",
								SourceType: loggingv1alpha1.LogSourceTypeApp,
								OutputRefs: []string{fluentDeployment.Name},
							},
							{
								Name:       "test-audit",
								SourceType: loggingv1alpha1.LogSourceTypeAudit,
								OutputRefs: []string{fluentDeployment.Name},
							},
							{
								Name:       "test-Infra",
								SourceType: loggingv1alpha1.LogSourceTypeInfra,
								OutputRefs: []string{fluentDeployment.Name},
							},
						},
					},
				}

				if err := e2e.CreateLogForwardingTechPreview(fw); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
				}

				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}
			})

			It("should send logs to the output", func() {
				name := fluentDeployment.GetName()
				Expect(e2e.LogStores[name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStores[name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStores[name].HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluent-receiver", "fluentd"})
		})
	})
})

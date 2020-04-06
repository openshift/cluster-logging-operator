package fluent

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LogForwarding", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err              error
		syslogDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		testDir          string
		forwarding       *logforward.LogForwarding
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
		testDir = filepath.Dir(filename)
	})
	Describe("when ClusterLogging is configured with 'forwarding' to an external syslog server", func() {

		BeforeEach(func() {
			forwarding = &logforward.LogForwarding{
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
							Name: helpers.SyslogReceiverName,
							Type: logforward.OutputTypeSyslog,
						},
					},
					Pipelines: []logforward.PipelineSpec{
						logforward.PipelineSpec{
							Name:       "test-infra",
							OutputRefs: []string{helpers.SyslogReceiverName},
							SourceType: logforward.LogSourceTypeInfra,
						},
					},
				},
			}
		})

		Context("with the new syslog plugin", func() {
			Context("and tcp receiver", func() {

				BeforeEach(func() {

					cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
				})

				Context("and with TLS disabled", func() {

					BeforeEach(func() {

						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						forwarding.Spec.Outputs[0].Endpoint = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					})
				})

				Context("and with TLS enabled", func() {

					BeforeEach(func() {

						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						forwarding.Spec.Outputs[0].Endpoint = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
						forwarding.Spec.Outputs[0].Secret = &logforward.OutputSecretSpec{
							Name: syslogDeployment.ObjectMeta.Name,
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
					})
				})
			})

			Context("and udp receiver", func() {

				Context("and TLS disabled", func() {

					BeforeEach(func() {

						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, false); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
						if err := e2e.CreateClusterLogging(cr); err != nil {
							Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
						}
						forwarding.Spec.Outputs[0].Endpoint = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					})
				})

				Context("and TLS enabled", func() {

					BeforeEach(func() {

						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, true); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
						if err := e2e.CreateClusterLogging(cr); err != nil {
							Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
						}
						forwarding.Spec.Outputs[0].Endpoint = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
						forwarding.Spec.Outputs[0].Secret = &logforward.OutputSecretSpec{
							Name: syslogDeployment.ObjectMeta.Name,
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
					})
				})
			})
		})

		Context("with the old syslog plugin", func() {

			Context("and tcp receiver", func() {

				BeforeEach(func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
					forwarding.Spec.Outputs[0].Endpoint = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarding.ObjectMeta.Annotations = map[string]string{
						k8shandler.UseOldRemoteSyslogPlugin: "enabled",
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
				})
			})

			Context("and udp receiver", func() {

				BeforeEach(func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, false); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
					forwarding.Spec.Outputs[0].Endpoint = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarding.ObjectMeta.Annotations = map[string]string{
						k8shandler.UseOldRemoteSyslogPlugin: "enabled",
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
				})
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion([]string{"fluentd", "syslog-receiver"})
		})

	})

})

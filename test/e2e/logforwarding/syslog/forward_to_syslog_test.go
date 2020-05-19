package fluent

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	grepprogname = `grep programname %s | head -n 1 | awk -F',' '{printf("%%s\n",$2)}' | awk -F\'  '{printf("%%s\n",$2)}'`
	grepappname  = `grep APP-NAME %s | head -n 1 | awk -F',' '{printf("%%s\n",$3)}' | awk -F\'  '{printf("%%s\n",$2)}'`
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
					DisableDefaultForwarding: true,
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

					DescribeTable("should send logs to the forward.Output logstore", func(rfc helpers.SyslogRfc) {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, rfc); err != nil {
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
						Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						Expect(e2e.LogStore.GrepLogs(grepprogname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected syslogtag to be \"fluentd\"")
						Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected APP-NAME to be \"fluentd\"")
					},
						//Entry("with rfc 3164", helpers.Rfc3164))
						Entry("with rfc 5424", helpers.Rfc5424))
				})

				Context("and with TLS enabled", func() {

					DescribeTable("should send logs to the forward.Output logstore", func(rfc helpers.SyslogRfc) {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, rfc); err != nil {
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
						Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						Expect(e2e.LogStore.GrepLogs(grepprogname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected syslogtag to be \"fluentd\"")
						Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected APP-NAME to be \"fluentd\"")
					},
						//Entry("with rfc 3164", helpers.Rfc3164))
						Entry("with rfc 5424", helpers.Rfc5424))
				})
			})

			Context("and udp receiver", func() {

				Context("and TLS disabled", func() {

					DescribeTable("should send logs to the forward.Output logstore", func(rfc helpers.SyslogRfc) {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, false, rfc); err != nil {
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
						Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						Expect(e2e.LogStore.GrepLogs(grepprogname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected syslogtag to be \"fluentd\"")
						Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected APP-NAME to be \"fluentd\"")
					},
						//Entry("with rfc 3164", helpers.Rfc3164))
						Entry("with rfc 5424", helpers.Rfc5424))
				})

				Context("and TLS enabled", func() {

					DescribeTable("should send logs to the forward.Output logstore", func(rfc helpers.SyslogRfc) {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, true, rfc); err != nil {
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
						Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						Expect(e2e.LogStore.GrepLogs(grepprogname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected syslogtag to be \"fluentd\"")
						Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal("fluentd"), "Expected APP-NAME to be \"fluentd\"")
					},
						//Entry("with rfc 3164", helpers.Rfc3164))
						Entry("with rfc 5424", helpers.Rfc5424))
				})
			})
		})

		Context("with the old syslog plugin", func() {

			Context("and tcp receiver", func() {

				DescribeTable("should send logs to the forward.Output logstore", func(rfc helpers.SyslogRfc) {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, rfc); err != nil {
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				},
					Entry("with rfc 3164", helpers.Rfc3164))
			})

			Context("and udp receiver", func() {

				DescribeTable("should send logs to the forward.Output logstore", func(rfc helpers.SyslogRfc) {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, false, rfc); err != nil {
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				},
					Entry("with rfc 3164", helpers.Rfc3164))
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion([]string{"fluentd", "syslog-receiver"})
		})

	})

})

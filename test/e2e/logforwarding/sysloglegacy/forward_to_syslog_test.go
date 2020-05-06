package sysloglegacy

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("ClusterLogForwarder", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err              error
		syslogDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		testDir          string
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
		testDir = filepath.Dir(filename)
	})
	Describe("when ClusterLogging is configured with 'forwarder' to an external syslog server", func() {

		Context("with the legacy syslog plugin", func() {

			Context("and tcp receiver", func() {

				BeforeEach(func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, helpers.Rfc3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					const conf = `
<store>
	@type syslog_buffered
	@id syslogid
	remote_syslog syslog-receiver.openshift-logging.svc
	port 24224
	hostname ${hostname}
	facility user
	severity debug
</store>
					`
					//create configmap syslog/"syslog.conf"
					if err = e2e.CreateLegacySyslogConfigMap(syslogDeployment.Namespace, conf); err != nil {
						Fail(fmt.Sprintf("Unable to create legacy syslog.conf configmap: %v", err))
					}

					components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
					cr := helpers.NewClusterLogging(components...)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
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
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, false, helpers.Rfc3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					const conf = `
<store>
	@type syslog
	@id syslogid
	remote_syslog syslog-receiver.openshift-logging.svc
	port 24224
	hostname ${hostname}
	facility user
	severity debug
</store>
					`
					//create configmap syslog/"syslog.conf"
					if err = e2e.CreateLegacySyslogConfigMap(syslogDeployment.Namespace, conf); err != nil {
						Fail(fmt.Sprintf("Unable to create legacy syslog.conf configmap: %v", err))
					}

					components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
					cr := helpers.NewClusterLogging(components...)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
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
			e2e.WaitForCleanupCompletion([]string{"fluentd", "syslog-receiver", "elasticsearch"})
		})

	})

})

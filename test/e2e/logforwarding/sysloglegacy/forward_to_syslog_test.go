package sysloglegacy

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("LogForwarding", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err              error
		syslogDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
	})
	Describe("when ClusterLogging is configured with 'forwarding' to an external syslog server", func() {

		Context("with the legacy syslog plugin", func() {

			Context("and tcp receiver", func() {

				BeforeEach(func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(corev1.ProtocolTCP); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					fluentdConfigMap := k8shandler.NewConfigMap(
						"syslog",
						syslogDeployment.Namespace,
						map[string]string{
							"syslog.conf": conf,
						},
					)
					if _, err = e2e.KubeClient.Core().ConfigMaps(syslogDeployment.Namespace).Create(fluentdConfigMap); err != nil {
						Fail(fmt.Sprintf("Unable to create legacy syslog.conf configmap: %v", err))
					}

					components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
					cr := helpers.NewClusterLogging(components...)
					cr.ObjectMeta.Annotations[k8shandler.ForwardingAnnotation] = "disabled"
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
		})

	})

})

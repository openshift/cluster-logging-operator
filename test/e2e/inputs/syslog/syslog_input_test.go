// go:build !fluentd
package syslog

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"time"
)

var _ = Describe("[ClusterLogForwarder] Inputs Syslog", func() {

	var (
		logGenerator    = "log-generator"
		forwarderName   = "my-log-collector"
		destinationHost string
		testNamespace   string
		socatCmd        string
		e2e             = framework.NewE2ETestFramework()
		receiver        *framework.VectorHttpReceiverLogStore
		err             error
		forwarder       *logging.ClusterLogForwarder
		destinationPort = 10514

		caFile  = "/etc/collector/syslog/tls.crt"
		keyFile = "/etc/collector/syslog/tls.key"
	)

	Describe("with vector collector", func() {
		BeforeEach(func() {
			// init the framework
			e2e = framework.NewE2ETestFramework()
			forwarder = testruntime.NewClusterLogForwarder()
			testNamespace = e2e.CreateTestNamespace()
			forwarder.Namespace = testNamespace
			forwarder.Name = forwarderName

			// deploy receiver
			receiver, err = e2e.DeployHttpReceiver(testNamespace)
			Expect(err).To(BeNil())
			sa, err := e2e.BuildAuthorizationFor(testNamespace, forwarderName).
				AllowClusterRole("collect-infrastructure-logs").
				Create()
			Expect(err).To(BeNil())
			forwarder.Spec.ServiceAccountName = sa.Name

			//deploy forwarder
			testruntime.NewClusterLogForwarderBuilder(forwarder).
				FromInputWithVisitor("syslog", func(spec *logging.InputSpec) {
					spec.Receiver = &logging.ReceiverSpec{
						Type: logging.ReceiverTypeSyslog,
						ReceiverTypeSpec: &logging.ReceiverTypeSpec{
							Syslog: &logging.SyslogReceiver{
								Port: int32(destinationPort),
							},
						},
					}

				}).ToOutputWithVisitor(func(spec *logging.OutputSpec) {
				spec.Type = logging.OutputTypeHttp
				spec.URL = receiver.ClusterLocalEndpoint()
			}, "my-output")
			if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
			}
			if err := e2e.WaitForDaemonSet(testNamespace, forwarderName); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
			}

			// deploy log generator
			options := framework.LogGeneratorOptions{
				ContainerCount: 1,
				Labels: map[string]string{
					"component": logGenerator,
				},
			}
			if err := e2e.DeploySocat(testNamespace, logGenerator, forwarderName, options); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}

			destinationHost = fmt.Sprintf("%s-syslog.%s.svc.cluster.local", forwarderName, testNamespace)
			socatCmd = fmt.Sprintf("socat openssl-connect:%s:%d,cafile=%s,cert=%s,key=%s -",
				destinationHost, destinationPort, caFile, caFile, keyFile)
		})

		It("send syslog to inputs receiver", func() {
			host := "acme.com"
			app := "mortal-combat"
			pid := 6868
			msg := "Choose Your Destiny"
			msgId := "ID7"
			rfc3164 := fmt.Sprintf("<30>Apr 26 15:14:30 %s %s[%d]: %s", host, app, pid, msg)
			rfc5425 := fmt.Sprintf("<39>1 2024-04-26T15:36:49.214+03:00 %s %s %d %s - %s", host, app, pid, msgId, msg)
			// send syslog message
			cmd := fmt.Sprintf("echo %q | %s", rfc3164, socatCmd)
			_, err := e2e.PodExec(testNamespace, logGenerator, logGenerator, []string{"/bin/sh", "-c", cmd})
			if err != nil {
				Fail(fmt.Sprintf("Error execution write command: %v", err))
			}
			cmd = fmt.Sprintf("echo %q | %s", rfc5425, socatCmd)
			_, err = e2e.PodExec(testNamespace, logGenerator, logGenerator, []string{"/bin/sh", "-c", cmd})
			if err != nil {
				Fail(fmt.Sprintf("Error execution write command: %v", err))
			}
			time.Sleep(5 * time.Second)
			logs, err := receiver.ListJournalLogs()
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(2))
			Expect(logs[0].LogType).To(BeEquivalentTo("infrastructure"))
			Expect(logs[0].Message).To(BeEquivalentTo(msg))
			Expect(logs[1].LogType).To(BeEquivalentTo("infrastructure"))
			Expect(logs[1].Message).To(BeEquivalentTo(msg))
		})
		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(testNamespace, []string{"test"})
		})
	})
})

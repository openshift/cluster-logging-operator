package syslog

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/ViaQ/logerr/v2/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// grep "pod_name"=>"<generator-pod-name>" <syslog-log-filename>
	rsyslogFormatStr = `grep %s %%s| grep pod_name | tail -n 1 | awk -F' ' '{print %s}'`
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger := log.NewLogger("e2e-logforwarding")
	logger.Info("Running ", "filename", filename)
	var (
		err              error
		syslogDeployment *apps.Deployment
		e2e              = framework.NewE2ETestFramework()
		testDir          string
		forwarder        *logging.ClusterLogForwarder
		generatorPayload map[string]string
		waitlogs         string
		grepappname      string
		grepprocid       string
		grepmsgid        string
		logGenPod        string
		logGenNS         string
	)
	BeforeEach(func() {
		generatorPayload = map[string]string{
			"msgcontent": "My life is my message",
		}
	})
	JustBeforeEach(func() {
		if logGenNS, logGenPod, err = e2e.DeployJsonLogGenerator(generatorPayload); err != nil {
			logger.Error(err, "unable to deploy log generator.")
		}
		logger.Info("log generator pod: ", "podname", logGenPod)
		testDir = filepath.Dir(filename)
		// wait for current log-generator's logs to appear in syslog
		waitlogs = fmt.Sprintf(`[ $(grep %s %%s |grep pod_name| wc -l) -gt 0 ]`, logGenPod)
		grepappname = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$4")
		grepprocid = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$5")
		grepmsgid = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$6")
	})
	Describe("when the output is a third-party managed syslog", func() {
		BeforeEach(func() {
			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
		})
		Describe("with rfc5424", func() {
			BeforeEach(func() {
				forwarder = &logging.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       logging.ClusterLogForwarderKind,
						APIVersion: logging.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logging.ClusterLogForwarderSpec{
						Outputs: []logging.OutputSpec{
							{
								Name: "syslogout",
								Type: "syslog",
								OutputTypeSpec: logging.OutputTypeSpec{
									Syslog: &logging.Syslog{
										Facility: "user",
										Severity: "debug",
										AppName:  "myapp",
										ProcID:   "myproc",
										MsgID:    "mymsg",
										RFC:      "RFC5424",
									},
								},
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{"syslogout"},
								InputRefs:  []string{"application"},
							},
						},
					},
				}
			})
			DescribeTable("should be able to send logs to syslog receiver", func(tls bool, protocol corev1.Protocol) {
				if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, protocol, tls, framework.RFC5424); err != nil {
					Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
				}
				if protocol == corev1.ProtocolTCP {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				} else {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				}

				if tls {
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}
				logStore := e2e.LogStores[syslogDeployment.GetName()]
				Expect(logStore.HasInfraStructureLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				_, _ = logStore.GrepLogs(waitlogs, framework.DefaultWaitForLogsTimeout)
				expectedAppName := forwarder.Spec.Outputs[0].Syslog.AppName
				Expect(logStore.GrepLogs(grepappname, framework.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
				expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
				Expect(logStore.GrepLogs(grepmsgid, framework.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
				expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
				Expect(logStore.GrepLogs(grepprocid, framework.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
			},
				Entry("with TLS disabled, with TCP", false, corev1.ProtocolTCP),
				Entry("with TLS disabled, with UDP", false, corev1.ProtocolUDP),
				Entry("with TLS enabled, with TCP", true, corev1.ProtocolTCP),
				Entry("with TLS enabled, with UDP", true, corev1.ProtocolUDP),
			)
		})
		Describe("with rfc3164", func() {
			BeforeEach(func() {
				forwarder = &logging.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       logging.ClusterLogForwarderKind,
						APIVersion: logging.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logging.ClusterLogForwarderSpec{
						Outputs: []logging.OutputSpec{
							{
								Name: "syslogout",
								Type: "syslog",
								OutputTypeSpec: logging.OutputTypeSpec{
									Syslog: &logging.Syslog{
										Facility: "user",
										Severity: "debug",
										RFC:      "RFC3164",
										Tag:      "mytag",
									},
								},
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{"syslogout"},
								InputRefs:  []string{"application"},
							},
						},
					},
				}
			})
			DescribeTable("should be able to send logs to syslog receiver", func(useOldPlugin bool, tls bool, protocol corev1.Protocol) {
				if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, protocol, tls, framework.RFC3164); err != nil {
					Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
				}
				if protocol == corev1.ProtocolTCP {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				} else {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				}

				if tls {
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
				}
				if useOldPlugin {
					forwarder.ObjectMeta.Annotations = map[string]string{
						k8shandler.UseOldRemoteSyslogPlugin: "enabled",
					}
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}
				logStore := e2e.LogStores[syslogDeployment.GetName()]
				Expect(logStore.HasInfraStructureLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				if !useOldPlugin {
					_, _ = logStore.GrepLogs(waitlogs, framework.DefaultWaitForLogsTimeout)
					expectedAppName := forwarder.Spec.Outputs[0].Syslog.Tag
					Expect(logStore.GrepLogs(grepappname, framework.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
				}
			},
				// old syslog plugin does not support TLS, so set false for tls
				Entry("with old syslog plugin, with TCP", true, false, corev1.ProtocolTCP),
				Entry("with old syslog plugin, with UDP", true, false, corev1.ProtocolUDP),
				// using new plugin
				Entry("with new syslog plugin, without TLS, with TCP", false, false, corev1.ProtocolTCP),
				Entry("with new syslog plugin, without TLS, with UDP", false, false, corev1.ProtocolTCP),
				Entry("with new syslog plugin, with TLS, with TCP", false, true, corev1.ProtocolTCP),
				Entry("with new syslog plugin, with TLS, with UDP", false, true, corev1.ProtocolTCP),
			)
		})
		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
			e2e.WaitForCleanupCompletion(constants.OpenshiftNS, []string{constants.CollectorName, "syslog-receiver"})
			generatorPayload = map[string]string{}
		})
	})
})

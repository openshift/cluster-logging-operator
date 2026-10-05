package syslog

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	helpersyslog "github.com/openshift/cluster-logging-operator/test/helpers/syslog"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	testruntimeobs "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[ClusterLogForwarder] Syslog UDP connection recovery", func() {
	var (
		err              error
		e2e              *framework.E2ETestFramework
		forwarder        *obs.ClusterLogForwarder
		forwarderName    = "my-forwarder"
		deployNS         string
		syslogDeployment *apps.Deployment
		serviceAccount   *corev1.ServiceAccount
	)

	Describe("with vector collector over UDP", func() {
		BeforeEach(func() {
			e2e = framework.NewE2ETestFramework()
			deployNS = e2e.Test.NS.Name

			By("Create log generator first so it starts producing logs")
			msg := rand.Word(1024)
			logGenerator := testruntime.NewLogGeneratorDeployment(deployNS, "log-generator", int32(60), 1*time.Second, string(msg))
			Expect(e2e.Test.Create(logGenerator)).To(Succeed(), "failed to create log generator")

			if serviceAccount, err = e2e.BuildAuthorizationFor(deployNS, forwarderName).
				AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
				AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
				AllowClusterRole(framework.ClusterRoleCollectAuditLogs).Create(); err != nil {
				Fail(err.Error())
			}

			By("Deploy syslog UDP receiver first")
			syslogDeployment, err = e2e.DeploySyslogReceiver(deployNS, corev1.ProtocolUDP, false, helpersyslog.RFC5424)
			Expect(err).To(BeNil(), "should successfully deploy syslog UDP receiver")

			By("Create forwarder with the syslog receiver running")
			forwarder = testruntimeobs.NewClusterLogForwarderBuilder(obsruntime.NewClusterLogForwarder(deployNS, forwarderName, runtime.Initialize), func(clf *obs.ClusterLogForwarder) {
				clf.Spec.Collector = &obs.CollectorSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				}
				clf.Spec.ServiceAccount.Name = serviceAccount.Name
			}).
				FromInput(obs.InputTypeApplication).
				ToSyslogOutput(obs.SyslogRFC5424, func(output *obs.OutputSpec) {
					output.Syslog = &obs.Syslog{
						URL: fmt.Sprintf("udp://%s.%s.svc:514", framework.SyslogReceiverName, deployNS),
						RFC: obs.SyslogRFC5424,
						Tuning: &obs.SyslogTuningSpec{
							DeliveryMode: obs.DeliveryModeAtLeastOnce,
						},
						Facility:   "local0",
						Enrichment: obs.EnrichmentTypeKubernetesMinimal,
						AppName:    `{.systemd.u.SYSLOG_IDENTIFIER||.log_type||"-"}`,
						ProcId:     `{.systemd.t.PID||"-"}`,
						MsgId:      `{.systemd.u.MESSAGE_ID||"-"}`,
					}
				}).End()

			if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
			}

			By("Waiting for the collector daemonset")
			if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
				Fail(err.Error())
			}
		})

		Context("should recover and send logs after syslog UDP receiver restarts", func() {
			var podDeleteTest = func() {
				By("Verify initial logs are flowing")
				logStore := e2e.LogStores[syslogDeployment.GetName()]
				Expect(logStore.HasApplicationLogs(time.Minute*2)).To(BeTrue(), "expected to collect application logs initially")

				ctx := context.TODO()
				labelSelector := fmt.Sprintf("%s=%s", constants.LabelK8sComponent, syslogDeployment.Name)

				By("Delete syslog receiver pod repeatedly to trigger ICMP error caching on vector's UDP socket")
				for i := 0; i < 10; i++ {
					err = e2e.KubeClient.CoreV1().Pods(deployNS).DeleteCollection(ctx,
						metav1.DeleteOptions{GracePeriodSeconds: utils.GetPtr[int64](0)},
						metav1.ListOptions{LabelSelector: labelSelector},
					)
					Expect(err).To(BeNil(), "should be able to delete syslog receiver pods")
					time.Sleep(5 * time.Second)
				}
				Expect(e2e.WaitForDeployment(deployNS, syslogDeployment.Name, time.Second, time.Second*30)).To(Succeed(), "replicas should become available ")

				By("Verify that logs resume flowing after recovery")
				Expect(logStore.HasApplicationLogs(1*time.Minute)).To(BeTrue(), "expected to collect application logs after receiver recovers")
				//Fail("I should never get here")
			}

			It("when using a ClusterIP service", func() {
				podDeleteTest()
			})

			It("when using a NodePort service", func() {
				Expect(e2e.Test.Delete(runtime.NewService(deployNS, framework.SyslogReceiverName))).To(Succeed(), "should delete syslog receiver service")
				By("Replacing the ClusterIP service with a NodePort service")
				_, err = e2e.CreateSyslogService(syslogDeployment, corev1.ProtocolUDP, corev1.ServiceTypeNodePort)
				Expect(err).To(BeNil(), "should successfully create NodePort service")
				podDeleteTest()
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
		})
	})
})

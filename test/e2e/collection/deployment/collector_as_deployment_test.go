package deployment

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	corev1 "k8s.io/api/core/v1"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	apps "k8s.io/api/apps/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test collector deployment type", func() {
	const (
		receiverPort = 8080
		receiverName = "http-audit"
	)
	var (
		err                  error
		forwarder            *obs.ClusterLogForwarder
		forwarderName        = "my-forwarder"
		deploymentAnnotation = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
		fluentDeployment     *apps.Deployment
		deployNS             string
		e2e                  = framework.NewE2ETestFramework()
		serviceAccount       *corev1.ServiceAccount
	)

	Describe("with vector collector", func() {
		BeforeEach(func() {
			deployNS = e2e.CreateTestNamespace()
			fluentDeployment, err = e2e.DeployFluentdReceiverWithConf(deployNS, true, framework.FluentConfHTTPWithTLS)
			Expect(err).To(BeNil())
			logStore := e2e.LogStores[fluentDeployment.GetName()]

			if serviceAccount, err = e2e.BuildAuthorizationFor(deployNS, forwarderName).
				AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).Create(); err != nil {
				Fail(err.Error())
			}

			httpReceiverServiceName := fmt.Sprintf("%s-%s", forwarderName, receiverName)
			httpReceiverEndpoint := fmt.Sprintf("https://%s.%s.svc.cluster.local:%d", httpReceiverServiceName, deployNS, receiverPort)

			if err = e2e.DeployCURLLogGeneratorWithNamespaceAndEndpoint(deployNS, httpReceiverEndpoint); err != nil {
				Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
			}

			forwarder = obsruntime.NewClusterLogForwarder(deployNS, forwarderName, runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
				clf.Spec.ServiceAccount.Name = serviceAccount.Name
				clf.Annotations = deploymentAnnotation
				clf.Spec.Inputs = []obs.InputSpec{
					{
						Name: receiverName,
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: receiverPort,
							HTTP: &obs.HTTPReceiver{
								Format: obs.HTTPReceiverFormatKubeApiAudit,
							},
						},
					},
				}
				clf.Spec.Outputs = []obs.OutputSpec{
					{
						Name: "httpout-audit",
						Type: obs.OutputTypeHTTP,
						HTTP: &obs.HTTP{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("%s/logs/audit", logStore.ClusterLocalEndpoint()),
							},
							Method: "POST",
						},
						TLS: &obs.OutputTLSSpec{
							InsecureSkipVerify: true,
							TLSSpec: obs.TLSSpec{
								CA: &obs.ValueReference{
									Key:        constants.TrustedCABundleKey,
									SecretName: framework.FluentdSecretName,
								},
							},
						},
					},
				}
				clf.Spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       "input-receiver-logs",
						OutputRefs: []string{"httpout-audit"},
						InputRefs:  []string{"http-audit"},
					},
				}
			})
		})

		It("should be a deployment when annotated and http receiver is the only input", func() {
			Expect(e2e.CreateObservabilityClusterLogForwarder(forwarder)).To(Succeed(), "Exp. to create instance of ClusterLogForwarder")
			Expect(e2e.WaitForDeployment(forwarder.Namespace, forwarder.Name, 5*time.Second, 3*time.Minute)).To(Succeed(), "Exp. the collector to deploy as a deployment")

			logStore := e2e.LogStores[fluentDeployment.GetName()]
			Expect(logStore.HasAuditLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "exp. to find some audit logs")

			collectedLogs, err := logStore.RetrieveLogs()
			Expect(err).To(BeNil())
			auditLogsStr, ok := collectedLogs["audit"]
			Expect(ok).To(BeTrue())
			auditLogs := map[string]interface{}{}
			err = json.Unmarshal([]byte(auditLogsStr), &auditLogs)
			Expect(err).To(BeNil(), auditLogsStr)

			message, ok := auditLogs["log_message"]
			Expect(ok).To(BeTrue())
			Expect(message).To(Not(BeEmpty()))
		})

		AfterEach(func() {
			e2e.Cleanup()
		})
	})

})

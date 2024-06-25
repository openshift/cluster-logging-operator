package deployment

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"

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
		logGenNS             string
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
			if logGenNS, err = e2e.DeployCURLLogGenerator(httpReceiverEndpoint); err != nil {
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
								Format: obs.HTTPReceiverFormatKubeAPIAudit,
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
								CA: &obs.ConfigMapOrSecretKey{
									Key: constants.TrustedCABundleKey,
									Secret: &corev1.LocalObjectReference{
										Name: framework.FluentdSecretName,
									},
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
			if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
			}
			if err := e2e.WaitForResourceCondition(forwarder.Namespace, "deployment", forwarder.Name, "", "{.status.availableReplicas}", 1,
				func(out string) (bool, error) {
					out = strings.TrimSpace(out)
					if out == "" {
						return false, nil
					}
					available, err := strconv.Atoi(out)
					if err != nil {
						return false, err
					}
					return available >= 1, nil
				}); err != nil {
				Fail(err.Error())
			}

			logStore := e2e.LogStores[fluentDeployment.GetName()]
			Expect(logStore.HasAuditLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "exp. to find some audit logs")

			collectedLogs, err := logStore.RetrieveLogs()
			Expect(err).To(BeNil())
			auditLogsStr, ok := collectedLogs["audit"]
			fmt.Printf("---- %s\n", auditLogsStr)
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
			e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
		})
	})

})

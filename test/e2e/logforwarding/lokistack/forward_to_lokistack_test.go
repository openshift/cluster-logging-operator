package lokistack

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("[ClusterLogForwarder] Forward to Lokistack", func() {
	var (
		err               error
		e2e               = framework.NewE2ETestFramework()
		forwarder         *obs.ClusterLogForwarder
		forwarderName     = "my-forwarder"
		deployNS          string
		logGenNS          string
		serviceAccount    *corev1.ServiceAccount
		lokiStackOut      *obs.OutputSpec
		lokistackReceiver *framework.LokistackLogStore
	)

	BeforeEach(func() {
		deployNS = e2e.CreateTestNamespace()

		if err = e2e.DeployMinio(); err != nil {
			Fail(err.Error())
		}
		if err = e2e.DeployLokiOperator(); err != nil {
			Fail(err.Error())
		}
		if lokistackReceiver, err = e2e.DeployLokistackInNamespace(deployNS); err != nil {
			Fail(err.Error())
		}

		if serviceAccount, err = e2e.BuildAuthorizationFor(deployNS, forwarderName).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
			AllowClusterRole(framework.ClusterRoleCollectAuditLogs).
			AllowClusterRole(framework.ClusterRoleAllLogsWriter).
			AllowClusterRole(framework.ClusterRoleAllLogsReader).Create(); err != nil {
			Fail(err.Error())
		}

		outputName := "lokistack-otlp"
		forwarder = obsruntime.NewClusterLogForwarder(deployNS, "my-forwarder", runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
			clf.Spec.ServiceAccount.Name = serviceAccount.Name
			clf.Annotations = map[string]string{constants.AnnotationOtlpOutputTechPreview: "true"}
			clf.Spec.Pipelines = append(clf.Spec.Pipelines, obs.PipelineSpec{
				Name:       "all-logs-pipeline",
				OutputRefs: []string{outputName},
				InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
			})
		})

		lokiStackOut = &obs.OutputSpec{
			Name: outputName,
			Type: obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Namespace: deployNS,
					Name:      "lokistack-dev",
				},
				Authentication: &obs.LokiStackAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromServiceAccount,
					},
				},
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:           "service-ca.crt",
						ConfigMapName: "openshift-service-ca.crt",
					},
				},
			},
		}

		// Deploy log generator
		logGenNS = e2e.CreateTestNamespaceWithPrefix("clo-test-loader")
		generatorOpt := framework.NewDefaultLogGeneratorOptions()
		generatorOpt.Count = -1
		if err = e2e.DeployLogGeneratorWithNamespaceName(logGenNS, "log-generator", generatorOpt); err != nil {
			Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
		}
	})

	It("should send logs to lokistack's OTLP endpoint when dataModel == Otel", func() {
		lokiStackOut.LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		found, err := lokistackReceiver.HasApplicationLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasInfrastructureLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})

	It("should send logs to lokistack when dataModel is not spec'd", func() {
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		found, err := lokistackReceiver.HasApplicationLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasInfrastructureLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})
	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
	})
})

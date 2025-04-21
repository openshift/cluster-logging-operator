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
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("[ClusterLogForwarder] Forward to Lokistack", func() {
	const (
		forwarderName = "my-forwarder"
		logGenName    = "log-generator"
	)
	var (
		err               error
		e2e               = framework.NewE2ETestFramework()
		forwarder         *obs.ClusterLogForwarder
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
			clf.Spec.Collector = &obs.CollectorSpec{
				Resources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("500m"),
					},
				},
			}
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
		if err = e2e.DeployLogGeneratorWithNamespaceName(logGenNS, logGenName, generatorOpt); err != nil {
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

		otlpKey := "k8s_container_name"
		res, err := lokistackReceiver.GetApplicationLogsByKeyValue(serviceAccount.Name, otlpKey, logGenName, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

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

		expression := `| json kubernetes_container_iostream="stdout"`
		res, err := lokistackReceiver.GetApplicationLogsWithPipeline(serviceAccount.Name, expression, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

		found, err = lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasInfrastructureLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})

	It("should send logs to lokistack with otel equivalent default labels when data model is viaq", func() {
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

		key := `k8s_pod_name`
		res, err := lokistackReceiver.GetApplicationLogsByKeyValue(serviceAccount.Name, key, logGenName, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

		// Check stream values here - len and contents
		stream := res[0].Stream
		Expect(len(stream)).To(Equal(10))
		wantStreamLabels := []string{
			"k8s_container_name",
			"k8s_namespace_name",
			"k8s_pod_name",
			"k8s_node_name",
			"kubernetes_container_name",
			"kubernetes_namespace_name",
			"kubernetes_pod_name",
			"kubernetes_host",
			"log_type",
			"openshift_log_type"}

		for _, key := range wantStreamLabels {
			_, ok := stream[key]
			Expect(ok).To(BeTrue())
		}
	})

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
	})
})

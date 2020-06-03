package kafka

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LogForwarding", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		e2e = helpers.NewE2ETestFramework()
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
	})
	Describe("when ClusterLogging is configured with 'forwarding' to an administrator managed Kafka", func() {

		Context("write app, audit and infra logs on a single topic", func() {
			BeforeEach(func() {
				topics := []string{kafka.DefaultTopic}
				if err := e2e.DeployKafkaReceiver(topics); err != nil {
					Fail(fmt.Sprintf("Unable to deploy kafka receiver: %v", err))
				}

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				forwarder := &loggingv1.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       loggingv1.ClusterLogForwarderKind,
						APIVersion: loggingv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: loggingv1.ClusterLogForwarderSpec{
						Outputs: []loggingv1.OutputSpec{
							{
								Name: kafka.DeploymentName,
								Type: loggingv1.OutputTypeKafka,
								URL: fmt.Sprintf(
									"tls://%s/%s",
									e2e.LogStore.ClusterLocalEndpoint(),
									kafka.DefaultTopic,
								),
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name:       "kafka-app",
								InputRefs:  []string{loggingv1.InputNameApplication},
								OutputRefs: []string{kafka.DeploymentName},
							},
							{
								Name:       "kafka-audit",
								InputRefs:  []string{loggingv1.InputNameAudit},
								OutputRefs: []string{kafka.DeploymentName},
							},
							{
								Name:       "kafka-infra",
								InputRefs:  []string{loggingv1.InputNameInfrastructure},
								OutputRefs: []string{kafka.DeploymentName},
							},
						},
					},
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of ClusterLogForwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})

			It("should send logs to the forward.Output logstore", func() {
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStore.HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})

			AfterEach(func() {
				e2e.Cleanup()
			})
		})

		Context("split app, audit and infra on different topics", func() {
			BeforeEach(func() {
				topics := []string{kafka.AppLogsTopic, kafka.AuditLogsTopic, kafka.InfraLogsTopic}
				if err := e2e.DeployKafkaReceiver(topics); err != nil {
					Fail(fmt.Sprintf("Unable to deploy kafka receiver: %v", err))
				}

				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				forwarder := &loggingv1.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       loggingv1.ClusterLogForwarderKind,
						APIVersion: loggingv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: loggingv1.ClusterLogForwarderSpec{
						Outputs: []loggingv1.OutputSpec{
							{
								Name: fmt.Sprintf("%s-app-out", kafka.DeploymentName),
								Type: loggingv1.OutputTypeKafka,
								URL:  fmt.Sprintf("tls://%s", e2e.LogStore.ClusterLocalEndpoint()),
								OutputTypeSpec: loggingv1.OutputTypeSpec{
									Kafka: &v1.Kafka{
										Topic: kafka.AppLogsTopic,
									},
								},
							},
							{
								Name: fmt.Sprintf("%s-audit-out", kafka.DeploymentName),
								Type: loggingv1.OutputTypeKafka,
								URL:  fmt.Sprintf("tls://%s", e2e.LogStore.ClusterLocalEndpoint()),
								OutputTypeSpec: loggingv1.OutputTypeSpec{
									Kafka: &v1.Kafka{
										Topic: kafka.AuditLogsTopic,
									},
								},
							},
							{
								Name: fmt.Sprintf("%s-infra-out", kafka.DeploymentName),
								Type: loggingv1.OutputTypeKafka,
								URL:  fmt.Sprintf("tls://%s", e2e.LogStore.ClusterLocalEndpoint()),
								OutputTypeSpec: loggingv1.OutputTypeSpec{
									Kafka: &v1.Kafka{
										Topic: kafka.InfraLogsTopic,
									},
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name:       "kafka-app",
								InputRefs:  []string{loggingv1.InputNameApplication},
								OutputRefs: []string{fmt.Sprintf("%s-app-out", kafka.DeploymentName)},
							},
							{
								Name:       "kafka-audit",
								InputRefs:  []string{loggingv1.InputNameAudit},
								OutputRefs: []string{fmt.Sprintf("%s-audit-out", kafka.DeploymentName)},
							},
							{
								Name:       "kafka-infra",
								InputRefs:  []string{loggingv1.InputNameInfrastructure},
								OutputRefs: []string{fmt.Sprintf("%s-infra-out", kafka.DeploymentName)},
							},
						},
					},
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of ClusterLogForwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})

			It("should send logs to the forward.Output logstore", func() {
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStore.HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})

			AfterEach(func() {
				e2e.Cleanup()
			})
		})
	})
})

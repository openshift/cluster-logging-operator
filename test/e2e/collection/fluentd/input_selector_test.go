package fluentd

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/ViaQ/logerr/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[Collection] InputSelector filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		err              error
		fluentDeployment *apps.Deployment
		app              *apps.StatefulSet
		e2e              = helpers.NewE2ETestFramework()
		rootDir          string
	)
	appNamespace1 := "application-ns1"
	appNamespace2 := "application-ns2"
	appLabels1 := map[string]string{"name": "app1", "env": "env1"}
	appLabels2 := map[string]string{"name": "app2", "env": "env2"}

	Describe("when CLF has input selectors to collect application logs", func() {
		Describe("from pods identified by labels", func() {
			BeforeEach(func() {
				if _, err = oc.Literal().From("oc create ns %s", appNamespace1).Run(); err != nil {
					Fail("failed to create namespace")
				}
				if _, err = oc.Literal().From("oc create ns %s", appNamespace2).Run(); err != nil {
					Fail("failed to create namespace")
				}
			})
			BeforeEach(func() {
				if err := e2e.DeployLogGeneratorWithNamespaceAndLabels(appNamespace1, appLabels1); err != nil {
					Fail(fmt.Sprintf("Timed out waiting for the log generator 1 to deploy: %v", err))
				}
				if err := e2e.DeployLogGeneratorWithNamespaceAndLabels(appNamespace2, appLabels2); err != nil {
					Fail(fmt.Sprintf("Timed out waiting for the log generator 2 to deploy: %v", err))
				}
				rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
				if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, false); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}
				topics := []string{kafka.AppLogsTopic, kafka.AuditLogsTopic, kafka.InfraLogsTopic}
				app, err = e2e.DeployKafkaReceiver(topics)
				if err != nil {
					Fail(fmt.Sprintf("Unable to deploy kafka receiver: %v", err))
				}

				forwarder := &logging.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       logging.ClusterLogForwarderKind,
						APIVersion: logging.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logging.ClusterLogForwarderSpec{
						Inputs: []logging.InputSpec{
							{
								Name: "application-logs1",
								Application: &logging.Application{
									Namespaces: []string{},
									Selector: &metav1.LabelSelector{
										MatchLabels: appLabels1,
									},
								},
							},
							{
								Name: "application-logs2",
								Application: &logging.Application{
									Namespaces: []string{},
									Selector: &metav1.LabelSelector{
										MatchLabels: appLabels2,
									},
								},
							},
						},
						Outputs: []logging.OutputSpec{
							{
								Name: fluentDeployment.GetName(),
								Type: logging.OutputTypeFluentdForward,
								URL:  fmt.Sprintf("tcp://%s.%s.svc:24224", fluentDeployment.GetName(), fluentDeployment.GetNamespace()),
							},
							{
								Name: fmt.Sprintf("%s-app-out", kafka.DeploymentName),
								Type: logging.OutputTypeKafka,
								URL:  fmt.Sprintf("tls://%s", e2e.LogStores[app.Name].ClusterLocalEndpoint()),
								OutputTypeSpec: logging.OutputTypeSpec{
									Kafka: &logging.Kafka{
										Topic: kafka.AppLogsTopic,
									},
								},
								Secret: &logging.OutputSecretSpec{
									Name: kafka.DeploymentName,
								},
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "fluent-app-logs1",
								InputRefs:  []string{"application-logs1"},
								OutputRefs: []string{fluentDeployment.GetName()},
							},
							{
								Name:       "kafka-app-logs2",
								InputRefs:  []string{"application-logs2"},
								OutputRefs: []string{fmt.Sprintf("%s-app-out", kafka.DeploymentName)},
							},
						},
					},
				}

				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
				}
				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				if err := e2e.WaitFor(helpers.ComponentTypeCollector); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
				}
			})

			It("should send logs from specific applications by using labels", func() {
				Expect(e2e.LogStores[fluentDeployment.GetName()].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs with ")

				logs, err := e2e.LogStores[fluentDeployment.GetName()].ApplicationLogs(helpers.DefaultWaitForLogsTimeout)
				Expect(err).To(BeNil(), fmt.Sprintf("Error fetching logs: %v", err))
				Expect(len(logs)).To(Not(Equal(0)), "There were no documents returned in the logs")

				// verify only appLabels1 logs appear in Application logs
				for _, msg := range logs {
					log.Info("Print", "msg", msg)
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("name", "app1"))
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("env", "env1"))
				}

				Expect(e2e.LogStores[app.Name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs with ")

				logs, err = e2e.LogStores[app.Name].ApplicationLogs(helpers.DefaultWaitForLogsTimeout)
				Expect(err).To(BeNil(), fmt.Sprintf("Error fetching logs: %v", err))
				Expect(len(logs)).To(Not(Equal(0)), "There were no documents returned in the logs")

				// verify only appLabels2 logs appear in Application logs
				for _, msg := range logs {
					log.Info("Print", "msg", msg)
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("name", "app2"))
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("env", "env2"))
				}
			})

			AfterEach(func() {
				if _, err = oc.Literal().From("oc delete ns %s", appNamespace1).Run(); err != nil {
					Fail("failed to delete namespace")
				}
				if _, err = oc.Literal().From("oc delete ns %s", appNamespace2).Run(); err != nil {
					Fail("failed to delete namespace")
				}
				e2e.WaitForCleanupCompletion(appNamespace1, []string{"test"})
				e2e.WaitForCleanupCompletion(appNamespace2, []string{"test"})
				e2e.Cleanup()
				e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluent-receiver", "fluentd"})
			}, helpers.DefaultCleanUpTimeout)

		})
		Describe("from pods identified by labels and namespaces", func() {
			BeforeEach(func() {
				if _, err = oc.Literal().From("oc create ns %s", appNamespace1).Run(); err != nil {
					Fail("failed to create namespace")
				}
				if _, err = oc.Literal().From("oc create ns %s", appNamespace2).Run(); err != nil {
					Fail("failed to create namespace")
				}
			})
			BeforeEach(func() {
				if err := e2e.DeployLogGeneratorWithNamespaceAndLabels(appNamespace1, appLabels1); err != nil {
					Fail(fmt.Sprintf("Timed out waiting for the log generator 1 to deploy: %v", err))
				}
				if err := e2e.DeployLogGeneratorWithNamespaceAndLabels(appNamespace2, appLabels1); err != nil {
					Fail(fmt.Sprintf("Timed out waiting for the log generator 2 to deploy: %v", err))
				}
				rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
				if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, false); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}

				forwarder := &logging.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       logging.ClusterLogForwarderKind,
						APIVersion: logging.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logging.ClusterLogForwarderSpec{
						Inputs: []logging.InputSpec{
							{
								Name: "application-logs",
								Application: &logging.Application{
									Namespaces: []string{appNamespace1},
									Selector: &metav1.LabelSelector{
										MatchLabels: appLabels1,
									},
								},
							},
						},
						Outputs: []logging.OutputSpec{
							{
								Name: fluentDeployment.GetName(),
								Type: logging.OutputTypeFluentdForward,
								URL:  fmt.Sprintf("tcp://%s.%s.svc:24224", fluentDeployment.GetName(), fluentDeployment.GetNamespace()),
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "fluent-app-logs",
								InputRefs:  []string{"application-logs"},
								OutputRefs: []string{fluentDeployment.GetName()},
							},
						},
					},
				}

				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
				}
				cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				if err := e2e.WaitFor(helpers.ComponentTypeCollector); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
				}
			})

			It("should send logs with labels name:app1 and env:env1 from namespace application-ns1 to fluentd only", func() {
				Expect(e2e.LogStores[fluentDeployment.GetName()].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs with ")

				logs, err := e2e.LogStores[fluentDeployment.GetName()].ApplicationLogs(helpers.DefaultWaitForLogsTimeout)
				Expect(err).To(BeNil(), fmt.Sprintf("Error fetching logs: %v", err))
				Expect(len(logs)).To(Not(Equal(0)), "There were no documents returned in the logs")

				// verify only appLabels1 logs appear in Application logs
				for _, msg := range logs {
					log.Info("Print", "msg", msg)
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("name", "app1"))
					Expect(msg.Kubernetes.NamespaceName).To(Equal(appNamespace1))
				}
			})

			AfterEach(func() {
				if _, err = oc.Literal().From("oc delete ns %s", appNamespace1).Run(); err != nil {
					Fail("failed to delete namespace")
				}
				if _, err = oc.Literal().From("oc delete ns %s", appNamespace2).Run(); err != nil {
					Fail("failed to delete namespace")
				}
				e2e.WaitForCleanupCompletion(appNamespace1, []string{"test"})
				e2e.WaitForCleanupCompletion(appNamespace2, []string{"test"})
				e2e.Cleanup()
				e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluent-receiver", "fluentd"})
			}, helpers.DefaultCleanUpTimeout)

		})

	})
})

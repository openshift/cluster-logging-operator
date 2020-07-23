package multiple_outputs

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	. "github.com/openshift/cluster-logging-operator/test/helpers"
	eologgingv1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)

	e2e := NewE2ETestFramework()

	var (
		err            error
		rootDir        string
		fluentRcv      *apps.Deployment
		elasticsearch  *eologgingv1.Elasticsearch
		pipelineSecret *corev1.Secret
		selectors      = []string{"elasticsearch", "fluent-receiver", "fluentd"}
	)

	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
	})

	Describe("when multiple outputs are configured", func() {

		Describe("and both are accepting logs", func() {

			BeforeEach(func() {
				fluentRcv, err = e2e.DeployFluentdReceiver(rootDir, false)
				if err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}

				elasticsearch, pipelineSecret, err = e2e.DeployAnElasticsearchCluster(rootDir)
				if err != nil {
					Fail(fmt.Sprintf("Unable to deploy an elasticsearch instance: %v", err))
				}

				cr := NewClusterLogging(ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}

				forwarder := newClusterLogForwarder(fluentRcv, elasticsearch, pipelineSecret)
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
				}

				components := []LogComponentType{ComponentTypeCollector, ComponentTypeStore}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})

			It("should send logs to the fluentd receiver and elasticsearch", func() {
				stores := []string{fluentRcv.GetName(), elasticsearch.GetName()}
				for _, name := range stores {
					Expect(e2e.LogStores[name].HasInfraStructureLogs(DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs in store %q", name)
					Expect(e2e.LogStores[name].HasApplicationLogs(DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs in store %q", name)
					Expect(e2e.LogStores[name].HasAuditLogs(DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs in store %q", name)
				}
			})
		})

		Describe("and one store is not available", func() {

			BeforeEach(func() {
				fluentRcv, err = e2e.DeployFluentdReceiver(rootDir, false)
				if err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}

				elasticsearch := &eologgingv1.Elasticsearch{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "err-elasticsearch",
						Namespace: constants.OpenshiftNS,
					},
				}
				e2e.LogStores[elasticsearch.GetClusterName()] = &ElasticLogStore{Framework: e2e}

				cr := NewClusterLogging(ComponentTypeCollector)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}

				forwarder := newClusterLogForwarder(fluentRcv, elasticsearch, nil)
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
				}

				components := []LogComponentType{ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})

			It("should send logs to the fluentd receiver only", func() {
				fluentd := fluentRcv.GetName()
				Expect(e2e.LogStores[fluentd].HasInfraStructureLogs(DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs in store %q", fluentd)
				Expect(e2e.LogStores[fluentd].HasApplicationLogs(DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs in store %q", fluentd)
				Expect(e2e.LogStores[fluentd].HasAuditLogs(DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs in store %q", fluentd)
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, selectors)
		})
	})
})

func newClusterLogForwarder(fluentRcv *apps.Deployment, elasticsearch *eologgingv1.Elasticsearch, pipelineSecret *corev1.Secret) *loggingv1.ClusterLogForwarder {
	fluentdOutput := loggingv1.OutputSpec{
		Name: fluentRcv.GetName(),
		Type: loggingv1.OutputTypeFluentdForward,
		URL:  fmt.Sprintf("%s.%s.svc:24224", fluentRcv.GetName(), fluentRcv.GetNamespace()),
	}

	elasticOutput := loggingv1.OutputSpec{
		Name: elasticsearch.GetName(),
		Type: loggingv1.OutputTypeElasticsearch,
		URL:  fmt.Sprintf("%s.%s.svc:9200", elasticsearch.GetName(), elasticsearch.GetNamespace()),
	}

	if pipelineSecret != nil {
		elasticOutput.Secret = &loggingv1.OutputSecretSpec{
			Name: pipelineSecret.GetName(),
		}
	}

	return &loggingv1.ClusterLogForwarder{
		TypeMeta: metav1.TypeMeta{
			Kind:       loggingv1.ClusterLogForwarderKind,
			APIVersion: loggingv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "instance",
		},
		Spec: loggingv1.ClusterLogForwarderSpec{
			Outputs: []loggingv1.OutputSpec{
				fluentdOutput,
				elasticOutput,
			},
			Pipelines: []loggingv1.PipelineSpec{
				{
					Name:       "fluent-app-logs",
					InputRefs:  []string{loggingv1.InputNameApplication},
					OutputRefs: []string{fluentdOutput.Name},
				},
				{
					Name:       "fluent-audit-logs",
					InputRefs:  []string{loggingv1.InputNameAudit},
					OutputRefs: []string{fluentdOutput.Name},
				},
				{
					Name:       "fluent-infra-logs",
					InputRefs:  []string{loggingv1.InputNameInfrastructure},
					OutputRefs: []string{fluentdOutput.Name},
				},
				{
					Name:       "elasticsearch-app-logs",
					InputRefs:  []string{loggingv1.InputNameApplication},
					OutputRefs: []string{elasticOutput.Name},
				},
				{
					Name:       "elasticsearch-audit-logs",
					InputRefs:  []string{loggingv1.InputNameAudit},
					OutputRefs: []string{elasticOutput.Name},
				},
				{
					Name:       "elasticsearch-infra-logs",
					InputRefs:  []string{loggingv1.InputNameInfrastructure},
					OutputRefs: []string{elasticOutput.Name},
				},
			},
		},
	}
}

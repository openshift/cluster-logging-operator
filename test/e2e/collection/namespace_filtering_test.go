package collection

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	"github.com/ViaQ/logerr/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[Collection] Namespace filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		err            error
		pipelineSecret *corev1.Secret
		elasticsearch  *elasticsearch.Elasticsearch
		e2e            = framework.NewE2ETestFramework()
		rootDir        string
	)
	var appclient1 *client.Test
	var appclient2 *client.Test

	BeforeEach(func() {
		appclient1 = client.NewTest()
		appclient2 = client.NewTest()
	})
	BeforeEach(func() {
		if err := e2e.DeployLogGeneratorWithNamespace(appclient1.NS.Name); err != nil {
			Fail(fmt.Sprintf("Timed out waiting for the log generator 1 to deploy: %v", err))
		}
		if err := e2e.DeployLogGeneratorWithNamespace(appclient2.NS.Name); err != nil {
			Fail(fmt.Sprintf("Timed out waiting for the log generator 2 to deploy: %v", err))
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "/")
		if elasticsearch, pipelineSecret, err = e2e.DeployAnElasticsearchCluster(rootDir); err != nil {
			Fail(fmt.Sprintf("Unable to deploy an elastic instance: %v", err))
		}

		forwarder := &logging.ClusterLogForwarder{
			TypeMeta: metav1.TypeMeta{
				Kind:       logging.ClusterLogForwarderKind,
				APIVersion: logging.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance",
			},
			Spec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "application-logs",
						Application: &logging.Application{
							Namespaces: []string{appclient1.NS.Name},
						},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: elasticsearch.Name,
						Secret: &logging.OutputSecretSpec{
							Name: pipelineSecret.ObjectMeta.Name,
						},
						Type: logging.OutputTypeElasticsearch,
						URL:  fmt.Sprintf("https://%s.%s.svc:9200", elasticsearch.Name, elasticsearch.Namespace),
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "test-app",
						OutputRefs: []string{elasticsearch.Name},
						InputRefs:  []string{"application-logs"},
					},
				},
			},
		}
		if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}
	})

	GetNamespaces := func() ([]string, error) {
		logs, err := e2e.LogStores[elasticsearch.GetName()].ApplicationLogs(framework.DefaultWaitForLogsTimeout)
		if err != nil {
			return nil, errors.New("Getting error in application logs")
		}
		val := len(logs)
		namespaceMap := make(map[string]bool)
		var namespaceList []string
		// Parse each document and extract namespace value
		if val > 0 {
			for _, document := range logs {
				namespace_value := document.Kubernetes.NamespaceName
				if _, found := namespaceMap[namespace_value]; !found {
					namespaceMap[namespace_value] = true
					namespaceList = append(namespaceList, namespace_value)
				}
			}
		}
		return namespaceList, nil
	}

	DescribeTable("Running collector tests",
		func(collectorType helpers.LogComponentType) {
			cr := helpers.NewClusterLogging(collectorType)
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			if err := e2e.WaitFor(collectorType); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", collectorType, err))
			}
			Expect(e2e.LogStores[elasticsearch.GetName()].HasApplicationLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")

			namespaceList, err := GetNamespaces()
			Expect(err).To(BeNil(), fmt.Sprintf("Error fetching logs: %v", err))
			Expect(namespaceList).NotTo(BeNil())
			Expect(len(namespaceList)).To(Equal(1))
			Expect(namespaceList[0]).To(Equal(appclient1.NS.Name))
		},
		Entry("for fluentd collector", helpers.ComponentTypeCollectorFluentd),
		Entry("for vector collector", helpers.ComponentTypeCollectorVector),
	)

	AfterEach(func() {
		appclient1.Close()
		appclient2.Close()
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(constants.OpenshiftNS, []string{constants.CollectorName, "elasticsearch"})
	}, framework.DefaultCleanUpTimeout)
})

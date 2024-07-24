package v59_to_60

import (
	"context"
	"embed"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e/operator"
	receivers "github.com/openshift/cluster-logging-operator/test/framework/e2e/receivers/elasticsearch"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

//go:embed *.yaml
var embededFile embed.FS

var _ = Describe("[upgrade] From stable-5.9", func() {
	const (
		operatorNamespaceRH = "openshift-operators-redhat"
		eoChannel           = "stable-5.8"
		eoPackageName       = "elasticsearch-operator"

		startingVersion = "stable-5.9"
		nextVersion     = "stable-6.0"
		nextCatalog     = "cluster-logging-catalog"
	)
	var (
		e2e      *framework.E2ETestFramework
		eo       *operator.OperatorDeployment
		clo      *operator.OperatorDeployment
		logStore *receivers.ManagedElasticsearch

		ReadClusterLogging = func(file string) *logging.ClusterLogging {
			content, err := embededFile.ReadFile(file)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", file, err))
			}
			cl := &logging.ClusterLogging{}
			test.MustUnmarshal(string(content), cl)
			return cl
		}
		ReadClusterLogForwarder = func(file string) *logging.ClusterLogForwarder {
			content, err := embededFile.ReadFile(file)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", file, err))
			}
			clf := &logging.ClusterLogForwarder{}
			test.MustUnmarshal(string(content), clf)
			return clf
		}

		beforeEach = func() {
			e2e = framework.NewE2ETestFramework()
			e2e.CreateNamespace(operatorNamespaceRH)
			eo = e2e.NewOperatorDeployment(operatorNamespaceRH, eoPackageName, eoChannel)
			clo = e2e.NewOperatorDeployment(constants.OpenshiftNS, operator.PackageNameClusterLogging, startingVersion)
			Expect(eo.Deploy()).To(Succeed())
			Expect(clo.Deploy()).To(Succeed())
		}

		verifyLogs = func(logType obs.InputType) {
			// check log count twice to ensure logs are flowing
			indexName := fmt.Sprintf("%s-000001", logType)
			indices, err := logStore.Indices()
			Expect(err).NotTo(HaveOccurred())
			index, exists := indices.Get(indexName)
			Expect(exists).To(BeTrue(), fmt.Sprintf("Expected to find the %s index", logType))
			totEntries := index.DocCount()
			time.Sleep(5 * time.Second)
			indices, err = logStore.Indices()
			Expect(err).NotTo(HaveOccurred())
			index, exists = indices.Get(indexName)
			Expect(exists).To(BeTrue(), fmt.Sprintf("Expected to find the %s index", logType))

			Expect(index.DocCount()).To(BeNumerically(">", totEntries), "Expect log collection to continue")
		}

		verifyInfrastructureLogs = func() {
			verifyLogs(obs.InputTypeInfrastructure)
		}
		verifyAuditLogs = func() {
			verifyLogs(obs.InputTypeAudit)
		}
	)

	DescribeTable("to stable-6.0 when RH managed elasticsearch log storage exists", func(clFile, clfFile string, verify func()) {
		beforeEach()
		// before
		if clfFile != "" {
			clf := ReadClusterLogForwarder(clFile)
			Expect(e2e.Create(clf)).To(Succeed())
		}
		cl := ReadClusterLogging(clFile)
		Expect(e2e.Create(cl)).To(Succeed())

		Expect(e2e.WaitFor(helpers.ComponentTypeReceiverElasticsearchRHManaged)).To(Succeed())
		Expect(e2e.WaitFor(helpers.ComponentTypeVisualization)).To(Succeed())
		Expect(e2e.WaitFor(helpers.ComponentTypeCollector)).To(Succeed())
		logStore = receivers.NewManagedElasticsearch(e2e)
		Expect(logStore.HasInfraStructureLogs(time.Minute)).To(BeTrue()) // expect logs > 0

		// upgrade
		// verify all components still exist even though managed EO is no longer supported
		Expect(clo.UpdateSubscription(nextVersion, nextCatalog, constants.OpenshiftNS)).To(Succeed(), "Exp. to update the operator to the next version")
		Expect(e2e.WaitFor(helpers.ComponentTypeReceiverElasticsearchRHManaged)).To(Succeed(), "Exp. RH managed storage to be available")
		Expect(e2e.WaitFor(helpers.ComponentTypeVisualization)).To(Succeed(), "Exp. RH managed visualization to be available")
		Expect(e2e.WaitFor(helpers.ComponentTypeCollector)).To(Succeed(), "Exp. log collection to be be functional")

		pods, err := e2e.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=vector",
		})
		Expect(err).NotTo(HaveOccurred(), "Exp. no issues retrieving pods")
		Expect(pods.Items).NotTo(BeEmpty(), "Exp. the collector to be upgraded to vector implementation")

		verify()
	},
		Entry("should upgrade a deployment of only a ClusterLogging instance", "cl_instance_elasticsearch_fluentd.yaml", "", verifyInfrastructureLogs),
		Entry("should upgrade a deployment of ClusterLogging and ClusterLogForwarder instance", "cl_instance_elasticsearch_fluentd.yaml", "clf_instance_default_with_audit.yaml", func() {
			verifyInfrastructureLogs()
			verifyAuditLogs()
		}),
	)

	AfterEach(func() {
		e2e.Cleanup()
	})

})

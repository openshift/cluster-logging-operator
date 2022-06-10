package normalization

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

var _ = Describe("[Functional][Normalization] flatten labels", func() {

	var (
		framework         *functional.CollectorFunctionalFramework
		pb                *functional.PipelineBuilder
		applicationLabels map[string]string
		expLabels         map[string]string

		expectFlattenLabels = func(kubemeta types.Kubernetes) {
			Expect(kubemeta.FlatLabels).To(Not(BeNil()), fmt.Sprintf("Expected to find the kubernetes.flat_labels key in %#v", kubemeta))
			initialLabels := framework.Labels
			for k, v := range initialLabels {
				//expLabels change because kubernetes_metadata plugin replaces dots with underscores
				expLabels[strings.ReplaceAll(k, ".", "_")] = v
			}
			expFlatLabels := []string{}
			for k, v := range expLabels {
				expFlatLabels = append(expFlatLabels, fmt.Sprintf("%s=%s", k, v))
			}
			sort.Strings(expFlatLabels)
			sort.Strings(kubemeta.FlatLabels)

			Expect(kubemeta.FlatLabels).To(Equal(expFlatLabels), fmt.Sprintf("Expected to find the kubernetes.flat_labels key in %#v", kubemeta))
		}
	)

	BeforeEach(func() {
		applicationLabels = map[string]string{
			"app.kubernetes.io/name":       "test",
			"app.kubernetes.io/instance":   "functionaltest",
			"app.kubernetes.io/version":    "123",
			"app.kubernetes.io/component":  "thecomponent",
			"app.kubernetes.io/part-of":    "clusterlogging",
			"app.kubernetes.io/managed-by": "clusterloggingoperator",
			"app.kubernetes.io/created-by": "anoperator",
		}

		framework = functional.NewCollectorFunctionalFramework()
		expLabels = map[string]string{}
		for k, v := range applicationLabels {
			framework.Labels[k] = v
			//expLabels change because kubernetes_metadata plugin replaces dots with underscores
			expLabels[strings.ReplaceAll(k, ".", "_")] = v
		}
		pb = functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("for ES output", func() {
		It("should create 'kubernetes.flat_labels' with an array of 'kubernetes.labels' and remove all but the exclusions", func() {

			pb.ToElasticSearchOutput()
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogs(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			//verify we removed all but common labels
			Expect(logs[0].Kubernetes.Labels).To(Equal(expLabels), fmt.Sprintf("Expect to find only common kubernetes.labels in %#v", logs[0]))

			//verify the new key exists
			expectFlattenLabels(logs[0].Kubernetes)
		})
	})

	Context("for non-ES output", func() {
		It("should not remove 'kubernetes.labels' and not add 'kubernetes.flat_labels'", func() {
			pb.ToKafkaOutput()
			secret := kafka.NewBrokerSecret(framework.Namespace)
			Expect(framework.Test.Client.Create(secret)).To(Succeed())
			Expect(framework.Deploy()).To(BeNil())

			if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
				for k, v := range framework.Labels {
					//expLabels change because kubernetes_metadata plugin replaces dots with underscores
					expLabels[strings.ReplaceAll(k, ".", "_")] = v
				}
			} else {
				tempLabels := map[string]string{}
				for k, v := range framework.Labels {
					//expLabels change because kubernetes_metadata plugin replaces dots with underscores
					tempLabels[strings.ReplaceAll(k, "_", ".")] = v
				}
				expLabels = tempLabels
			}

			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeKafka)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			//verify the label key exists as-is
			Expect(logs[0].Kubernetes.Labels).To(Equal(expLabels), fmt.Sprintf("Expect to find every label in kubernetes.labels in %#v", logs[0]))

			if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
				//verify the new key exists
				expectFlattenLabels(logs[0].Kubernetes)
			}
		})

	})
})

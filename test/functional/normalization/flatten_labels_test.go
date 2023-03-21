//go:build fluentd || vector
// +build fluentd vector

package normalization

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"regexp"
	"sort"
	"time"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("[Functional][Normalization] flatten labels", func() {

	var (
		framework         *functional.CollectorFunctionalFramework
		pb                *functional.PipelineBuilder
		applicationLabels map[string]string
		expLabels         map[string]string

		dedotMatcher        = regexp.MustCompile("[./]")
		expectFlattenLabels = func(kubemeta types.Kubernetes) {
			Expect(kubemeta.FlatLabels).To(Not(BeNil()), fmt.Sprintf("Expected to find the kubernetes.flat_labels key in %#v", kubemeta))
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

		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)

		err := wait.PollImmediate(5*time.Second, 60*time.Second, func() (done bool, err error) {
			ns := framework.Test.NS
			if err := framework.Test.Client.Get(ns); err != nil {
				log.V(0).Error(err, "Failed trying to get NS to update labels")
				return false, nil
			}
			namespaceLabels := map[string]string{
				"foo":                                "bar",
				"foo.bar":                            "foobar",
				"foo.bar/baz":                        "foobarbaz",
				"pod-security.kubernetes.io/enforce": "privileged",
				"security.openshift.io/scc.podSecurityLabelSync": "false",
			}
			ns.Labels = namespaceLabels
			if err := framework.Test.Client.Update(ns); err != nil {
				log.V(0).Error(err, "Failed trying to update NS labels")
				return false, nil
			}
			return true, nil
		})
		Expect(err).To(BeNil(), "Unable to update NS labels")

		// Other kubernetes labels
		expLabels = map[string]string{}
		framework.Labels["app"] = "bar"
		for k, v := range applicationLabels {
			framework.Labels[k] = v
			expLabels[dedotMatcher.ReplaceAllString(k, "_")] = v
		}

		pb = functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("for ES output", func() {
		It("should create 'kubernetes.flat_labels' with an array of 'kubernetes.labels' and remove from 'kubernetes.labels' all but the 'common' ones", func() {

			pb.ToElasticSearchOutput()
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			// verify we removed all but common labels
			// This also verifies if the labels have been de-dotted
			Expect(logs[0].Kubernetes.Labels).To(Equal(expLabels), fmt.Sprintf("Expect to find only common kubernetes.labels in %#v", logs[0]))

			// Verify de-dotting process for namespace labels
			Expect(logs[0].Kubernetes.NamespaceLabels).To(HaveKey("foo"), "Expect 'foo' to not change")
			Expect(logs[0].Kubernetes.NamespaceLabels).To(HaveKey("foo_bar"), "Expect 'foo.bar' to be de-dotted to 'foo_bar'")
			Expect(logs[0].Kubernetes.NamespaceLabels).To(HaveKey("foo_bar_baz"), "Expect 'foo.bar/baz' to be de-dotted to 'foo_bar_baz'")

			for k, v := range framework.Test.Client.Labels {
				expLabels[dedotMatcher.ReplaceAllString(k, "_")] = v
			}
			for k, v := range framework.Labels {
				expLabels[dedotMatcher.ReplaceAllString(k, "_")] = v
			}

			//verify the new key exists
			expectFlattenLabels(logs[0].Kubernetes)
		})
	})

	Context("for non-ES output", func() {
		It("should not remove 'kubernetes.labels' and not add 'kubernetes.flat_labels'", func() {

			for k, v := range framework.Test.Client.Labels {
				expLabels[dedotMatcher.ReplaceAllString(k, "_")] = v
			}
			for k, v := range framework.Labels {
				expLabels[dedotMatcher.ReplaceAllString(k, "_")] = v
			}

			pb.ToKafkaOutput()
			framework.Secrets = append(framework.Secrets, kafka.NewBrokerSecret(framework.Namespace))
			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeKafka)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			//verify the label key exists as-is
			Expect(logs[0].Kubernetes.Labels).To(Equal(expLabels), fmt.Sprintf("Expect to find every label in kubernetes.labels in %#v", logs[0]))

			// Verify de-dotting process for namespace labels
			Expect(logs[0].Kubernetes.NamespaceLabels).To(HaveKey("foo"), "Expect 'foo' to not change")
			Expect(logs[0].Kubernetes.NamespaceLabels).To(HaveKey("foo_bar"), "Expect 'foo.bar' to be de-dotted to 'foo_bar'")
			Expect(logs[0].Kubernetes.NamespaceLabels).To(HaveKey("foo_bar_baz"), "Expect 'foo.bar/baz' to be de-dotted to 'foo_bar_baz'")

			if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
				//verify the new key exists
				expectFlattenLabels(logs[0].Kubernetes)
			} else {
				Expect(logs[0].Kubernetes.FlatLabels).To(BeEmpty(), "labels should not be flattened when using vector")
			}
		})

	})
})

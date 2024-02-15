package prune

import (
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

var _ = Describe("[Functional][Filters][Prune] Prune filter", func() {
	const (
		pruneFilterName = "my-prune"
	)

	var (
		f *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		f.Cleanup()
	})

	Describe("when prune filter is spec'd", func() {
		It("should prune logs of fields not defined in `NotIn` first and then prune fields defined in `In`", func() {
			f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
			specialCharLabel := "foo-bar/baz"
			f.Labels = map[string]string{specialCharLabel: "specialCharLabel"}
			f.Forwarder.Spec.Filters = []logging.FilterSpec{
				{
					Name: pruneFilterName,
					Type: logging.FilterPrune,
					FilterTypeSpec: logging.FilterTypeSpec{
						PruneFilterSpec: &logging.PruneFilterSpec{
							In:    []string{".kubernetes.namespace_name", ".kubernetes.container_name", `.kubernetes.labels."foo-bar/baz"`},
							NotIn: []string{".log_type", ".message", ".kubernetes", ".openshift", `."@timestamp"`},
						},
					},
				},
			}
			functional.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(logging.InputNameApplication).
				ToElasticSearchOutput()

			f.Forwarder.Spec.Pipelines = []logging.PipelineSpec{
				{
					Name:       "myPrunePipeline",
					FilterRefs: []string{pruneFilterName},
					InputRefs:  []string{logging.InputNameApplication, logging.InputNameAudit, logging.InputNameInfrastructure},
					OutputRefs: []string{logging.OutputTypeElasticsearch},
				},
			}

			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "my error message")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			logs, err := f.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeElasticsearch)

			log := logs[0]

			Expect(log.ViaQCommon.Message).ToNot(BeNil())
			Expect(log.ViaQCommon.LogType).ToNot(BeNil())
			Expect(log.Kubernetes).ToNot(BeNil())
			Expect(log.Openshift).ToNot(BeNil())
			Expect(log.Timestamp).ToNot(BeNil())
			Expect(log.Kubernetes.Annotations).ToNot(BeNil())
			Expect(log.Kubernetes.PodName).ToNot(BeNil())
			Expect(log.Kubernetes.FlatLabels).To(HaveLen(1))
			Expect(log.Kubernetes.FlatLabels).ToNot(ContainElement("foo-bar_baz=specialCharLabel"))

			Expect(log.Kubernetes.ContainerName).To(BeEmpty())
			Expect(log.Kubernetes.NamespaceName).To(BeEmpty())
			Expect(log.Level).To(BeEmpty())

		})

	})
})

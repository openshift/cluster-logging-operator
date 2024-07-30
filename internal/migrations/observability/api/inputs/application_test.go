package inputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapApplicationInputs", func() {
	Context("helper functions", func() {
		It("should map list of namespaces to a list of NamespaceContainerSpec", func() {
			namespaces := []string{"foo", "bar", "baz"}

			expObsApp := []obs.NamespaceContainerSpec{
				{
					Namespace: "foo",
				},
				{
					Namespace: "bar",
				},
				{
					Namespace: "baz",
				},
			}
			Expect(mapNamespacesToNamespacedContainers(namespaces)).To(Equal(expObsApp))
		})
		It("should map logging.NamespaceContainerSpec to observability.NamespaceContainerSpec", func() {
			loggingNSCont := []logging.NamespaceContainerSpec{
				{
					Namespace: "foo",
					Container: "bar",
				},
				{
					Namespace: "baz",
					Container: "foobar",
				},
			}

			expObsNSCont := []obs.NamespaceContainerSpec{
				{
					Namespace: "foo",
					Container: "bar",
				},
				{
					Namespace: "baz",
					Container: "foobar",
				},
			}

			Expect(mapNamespacedContainers(loggingNSCont)).To(Equal(expObsNSCont))
		})
	})

	DescribeTable("inputs", func(visit func(appSpec *logging.Application), expObsApp *obs.Application) {
		loggingApp := &logging.Application{}
		visit(loggingApp)
		Expect(MapApplicationInput(loggingApp)).To(Equal(expObsApp))
	},
		Entry("should map logging.Namespaces to obs.NamespaceContainer", func(appSpec *logging.Application) {
			appSpec.Namespaces = []string{"foo", "baz"}
		}, &obs.Application{
			Includes: []obs.NamespaceContainerSpec{
				{
					Namespace: "foo",
				},
				{
					Namespace: "baz",
				},
			},
		}),
		Entry("should map and combine logging.Namespaces && logging.Includes to obs.NamespaceContainer", func(appSpec *logging.Application) {
			appSpec.Namespaces = []string{"foo"}
			appSpec.Includes = []logging.NamespaceContainerSpec{
				{
					Namespace: "bar",
					Container: "bar-container",
				},
				{
					Namespace: "baz",
					Container: "baz-container",
				},
			}
		}, &obs.Application{
			Includes: []obs.NamespaceContainerSpec{
				{
					Namespace: "bar",
					Container: "bar-container",
				},
				{
					Namespace: "baz",
					Container: "baz-container",
				},
				{
					Namespace: "foo",
				},
			},
		}),
	)

	It("should map logging.Application to observability.Application", func() {
		loggingApp := logging.Application{
			Selector: &logging.LabelSelector{
				MatchLabels: map[string]string{},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "foo",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"bar"},
					},
				},
			},
			Includes: []logging.NamespaceContainerSpec{
				{
					Namespace: "foo",
					Container: "bar",
				},
			},
			Excludes: []logging.NamespaceContainerSpec{
				{
					Namespace: "bz",
					Container: "fz",
				},
			},
			ContainerLimit: &logging.LimitSpec{
				MaxRecordsPerSecond: 1000,
			},
		}

		expObsApp := &obs.Application{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "foo",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"bar"},
					},
				},
			},
			Includes: []obs.NamespaceContainerSpec{
				{
					Namespace: "foo",
					Container: "bar",
				},
			},
			Excludes: []obs.NamespaceContainerSpec{
				{
					Namespace: "bz",
					Container: "fz",
				},
			},
			Tuning: &obs.ContainerInputTuningSpec{
				RateLimitPerContainer: &obs.LimitSpec{
					MaxRecordsPerSecond: 1000,
				},
			},
		}
		Expect(MapApplicationInput(&loggingApp)).To(Equal(expObsApp))
	})
})

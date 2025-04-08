package runtime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/version"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Object", func() {
	var (
		nsFoo = runtime.NewNamespace("foo")
	)

	It("generates ID", func() {
		var o runtime.Object
		o = runtime.NewNamespace("foo")
		Expect(runtime.ID(o)).To(Equal("/v1/namespaces/foo"))
		o = runtime.NewDaemonSet("foo", "name")
		Expect(runtime.ID(o)).To(Equal("apps/v1/namespaces/foo/daemonsets/name"))
		o = &corev1.PodList{}
		Expect(runtime.ID(o)).To(Equal("/v1, Kind=PodList"))
	})

	DescribeTable("Decode",
		func(manifest string, o runtime.Object) {
			got := runtime.Decode(manifest)
			Expect(got).To(EqualDiff(o), "%#v", manifest)
		},
		Entry("YAML string ns", test.YAMLString(nsFoo), nsFoo),
		Entry("JSON string ns", test.JSONLine(nsFoo), nsFoo),
	)

	DescribeTable("#Labels.Includes",
		func(success bool, matches map[string]string) {
			obj := runtime.NewNamespace("test")
			obj.Labels = map[string]string{
				"foo":                       "bar",
				constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
				constants.LabelK8sVersion:   version.Version,
			}
			Expect(runtime.Labels(obj).Includes(matches)).To(Equal(success))
		},
		Entry("matches a subset", true, map[string]string{
			constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
			constants.LabelK8sVersion:   version.Version,
		}),
		Entry("matches entire set", true, map[string]string{
			"foo":                       "bar",
			constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
			constants.LabelK8sVersion:   version.Version,
		}),
		Entry("fails a superset", false, map[string]string{
			"foo":                       "bar",
			constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
			constants.LabelK8sVersion:   version.Version,
			"xyz":                       "abc",
		}),
		Entry("fails a subset", false, map[string]string{
			constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
			constants.LabelK8sVersion:   "someother",
		}),
	)

	It("panics on bad manifest string", func() {
		Expect(func() { _ = runtime.Decode("bad manifest") }).To(Panic())
	})

	DescribeTable("New",
		func(got, want runtime.Object) { Expect(got).To(EqualDiff(want)) },
		Entry("NewNamespace", runtime.NewNamespace("foo"), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		}),
		Entry("NewConfigMap", runtime.NewConfigMap("ns", "foo", nil), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "ns"},
			TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			Data:       map[string]string{},
		}),
		Entry("NewServiceMonitor", runtime.NewServiceMonitor("ns", "foo"), &monitoringv1.ServiceMonitor{
			ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "ns"},
			TypeMeta:   metav1.TypeMeta{Kind: "ServiceMonitor", APIVersion: "monitoring.coreos.com/v1"},
		}),
	)
})

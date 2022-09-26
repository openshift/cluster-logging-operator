package runtime

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Object", func() {
	var (
		nsFoo = NewNamespace("foo")
	)

	It("generates ID", func() {
		var o Object
		o = NewNamespace("foo")
		Expect(ID(o)).To(Equal("/v1/namespaces/foo"))
		o = NewDaemonSet("foo", "name")
		Expect(ID(o)).To(Equal("apps/v1/namespaces/foo/daemonsets/name"))
		o = &corev1.PodList{}
		Expect(ID(o)).To(Equal("/v1, Kind=PodList"))
	})

	DescribeTable("Decode",
		func(manifest string, o Object) {
			got := Decode(manifest)
			Expect(got).To(EqualDiff(o), "%#v", manifest)
		},
		Entry("YAML string ns", test.YAMLString(nsFoo), nsFoo),
		Entry("JSON string ns", test.JSONLine(nsFoo), nsFoo),
	)

	It("panics on bad manifest string", func() {
		Expect(func() { _ = Decode("bad manifest") }).To(Panic())
	})

	DescribeTable("New",
		func(got, want Object) { Expect(got).To(EqualDiff(want)) },
		Entry("NewNamespace", NewNamespace("foo"), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		}),
		Entry("NewConfigMap", NewConfigMap("ns", "foo", nil), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "ns"},
			TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			Data:       map[string]string{},
		}),
		Entry("NewServiceMonitor", NewServiceMonitor("ns", "foo"), &monitoringv1.ServiceMonitor{
			ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "ns"},
			TypeMeta:   metav1.TypeMeta{Kind: "ServiceMonitor", APIVersion: "monitoring.coreos.com/v1"},
		}),
	)
})

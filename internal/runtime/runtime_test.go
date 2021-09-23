package runtime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	runtime2 "github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Object", func() {
	var (
		nsFoo = runtime.NewNamespace("foo")
		clf   = runtime2.NewClusterLogForwarder()
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
		Entry("YAML string clf", test.YAMLString(clf), clf),
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
	)
})

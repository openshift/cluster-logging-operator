package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	. "github.com/openshift/cluster-logging-operator/test/client"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Client", func() {
	var (
		t    *Test
		data map[string]string
	)

	BeforeEach(func() {
		t = NewTest()
		data = map[string]string{"a": "b"}
	})

	AfterEach(func() { t.Close() })

	It("creates object with data and automatic Labels", func() {
		cm := runtime.NewConfigMap(t.NS.Name, "foo", data)
		ExpectOK(t.Create(cm))
		cm2 := runtime.NewConfigMap(t.NS.Name, "foo", nil)
		ExpectOK(t.Get(cm2))
		Expect(cm2.Data).To(Equal(data))
		Expect(cm2.Labels).To(HaveKeyWithValue(LabelKey, LabelValue))
	})

	It("re-creates existing namespace", func() {
		// Namespaces don't delete synchronously, so an important case.
		ExpectOK(t.Recreate(t.NS))
	})

	It("re-creates existing object", func() {
		cm := runtime.NewConfigMap(t.NS.Name, "foo", map[string]string{"a": "b"})
		ExpectOK(t.Create(cm))
		cm2 := runtime.NewConfigMap(t.NS.Name, "foo", map[string]string{"x": "y"})
		ExpectOK(t.Recreate(cm2))
		cm3 := runtime.NewConfigMap(t.NS.Name, "foo", nil)
		ExpectOK(t.Get(cm3))
		Expect(cm3.Data).To(Equal(map[string]string{"x": "y"}))
		ExpectOK(t.Recreate(t.NS))
	})

	It("creates non-existing object", func() {
		cm := runtime.NewConfigMap(t.NS.Name, "foo", data)
		ExpectOK(t.Recreate(cm))
		cm2 := runtime.NewConfigMap(t.NS.Name, "foo", nil)
		ExpectOK(t.Get(cm2))
		Expect(cm2.Data).To(Equal(data))
	})

	It("lists objects", func() {
		l := &corev1.NodeList{}
		ExpectOK(t.List(l))
		Expect(l.Items).NotTo(BeEmpty())
	})
})

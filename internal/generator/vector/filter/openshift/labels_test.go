package openshift_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift"
)

var _ = Describe("", func() {
	It("should return no filter when the label set is nil", func() {
		Expect(openshift.NewLabelsFilter(nil).VRL()).To(BeEmpty())
	})
	It("should return no filter when the label set is empty", func() {
		Expect(openshift.NewLabelsFilter(map[string]string{}).VRL()).To(BeEmpty())
	})
	It("should return a filter with that will set the desired labels", func() {
		Expect(openshift.NewLabelsFilter(map[string]string{
			"foo": "bar",
			"xyz": "abc",
		}).VRL()).To(Equal(`._internal.openshift.labels = .openshift.labels = {"foo":"bar","xyz":"abc"}`))
	})
})

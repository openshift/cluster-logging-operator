package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
)

type handler struct{ what interface{} }

func (h *handler) ElasticSearch(o *outputs.ElasticSearch) error { h.what = o; return nil }
func (h *handler) FluentForward(o *outputs.FluentForward) error { h.what = o; return nil }
func (h *handler) Syslog(o *outputs.Syslog) error               { h.what = o; return nil }

var _ OutputTypeHandler = (*handler)(nil)

var _ = Describe("OutputSpec", func() {
	DescribeTable("recognizes valid type names",
		func(name string, value interface{}) {
			spec := OutputSpec{Type: name}
			var h handler
			Expect(spec.HandleType(&h)).To(Succeed())
			Expect(h.what).To(Equal(value))
		},
		Entry("syslog type", "syslog", (*outputs.Syslog)(nil)),
		Entry("elasticsearch type", OutputTypeElasticsearch, (*outputs.ElasticSearch)(nil)),
		Entry("fluentforward type", OutputTypeFluentForward, (*outputs.FluentForward)(nil)),
	)
	It("gets syslog spec", func() {
		spec := OutputSpec{
			Type:           "syslog",
			OutputTypeSpec: OutputTypeSpec{Syslog: &outputs.Syslog{Severity: "Critical"}},
		}
		var h handler
		Expect(spec.HandleType(&h)).To(Succeed())
		Expect(h.what).To(Equal(spec.Syslog))
	})
	It("gets fluentForward spec", func() {
		spec := OutputSpec{
			Type:           "fluentForward",
			OutputTypeSpec: OutputTypeSpec{FluentForward: &outputs.FluentForward{}},
		}
		var h handler
		Expect(spec.HandleType(&h)).To(Succeed())
		Expect(h.what).To(Equal(spec.FluentForward))
	})
	It("rejects mismatched type and spec", func() {
		spec := OutputSpec{
			Type:           "syslog",
			OutputTypeSpec: OutputTypeSpec{FluentForward: &outputs.FluentForward{}},
		}
		var h handler
		Expect(spec.HandleType(&h)).To(Succeed())
		Expect(h.what).To(Equal(spec.Syslog))
	})
	It("rejects empty/missing type", func() {
		spec := OutputSpec{}
		var h handler
		Expect(spec.HandleType(&h)).To(MatchError("not a valid output type: ''"))
	})
	It("rejects unknown type", func() {
		spec := OutputSpec{Type: "bad"}
		var h handler
		Expect(spec.HandleType(&h)).To(MatchError("not a valid output type: 'bad'"))
	})
})

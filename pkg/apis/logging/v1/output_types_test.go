package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

type handler struct{ what interface{} }

func (h *handler) ElasticSearch(o *Elasticsearch) error { h.what = o; return nil }
func (h *handler) FluentForward(o *FluentForward) error { h.what = o; return nil }
func (h *handler) Syslog(o *Syslog) error               { h.what = o; return nil }

var _ = Describe("OutputSpec", func() {
	It("recognizes valid type names", func() {
		for _, s := range []string{OutputTypeElasticsearch, OutputTypeFluentForward, OutputTypeSyslog} {
			Expect(IsOutputTypeName(s)).To(BeTrue(), "expect recognize %s", s)
		}
	})
	It("rejects unknown type", func() {
		Expect(IsOutputTypeName("bad")).To(BeFalse())
	})
})

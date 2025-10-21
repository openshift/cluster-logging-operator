package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("[internal][validations] ClusterLogForwarder will validate headers in Elasticsearch Output", func() {
	var (
		es   *v1.Elasticsearch
		spec v1.OutputSpec
	)
	BeforeEach(func() {
		es = &v1.Elasticsearch{}
		spec = v1.OutputSpec{
			Name:          "esOutput",
			Type:          v1.OutputTypeElasticsearch,
			Elasticsearch: es,
		}
	})

	Context("#validateElasticsearchHeaders", func() {

		It("should pass validation with empty headers", func() {
			Expect(validateElasticsearchHeaders(spec)).To(BeEmpty())
		})
		It("should pass validation when no invalid headers set", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Accept": "application/json",
			}
			Expect(validateElasticsearchHeaders(spec)).To(BeEmpty())
		})
		It("should fail validation when the Content-Type header is set", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Content-Type": "application/json",
			}
			Expect(validateElasticsearchHeaders(spec)).ToNot(BeEmpty())
		})
		It("should fail validation when the Authorization header is set", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Authorization": "test",
			}
			Expect(validateElasticsearchHeaders(spec)).ToNot(BeEmpty())
		})
		It("should pass validation when no Elasticsearch Output", func() {
			spec = v1.OutputSpec{
				Name:          "esOutput",
				Type:          v1.OutputTypeElasticsearch,
				Elasticsearch: &v1.Elasticsearch{},
			}
			Expect(validateElasticsearchHeaders(spec)).To(BeEmpty())
		})
	})
})

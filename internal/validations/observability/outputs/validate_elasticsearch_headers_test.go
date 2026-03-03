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

		It("should fail validation when duplicate case-variant headers with same value", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Accept": "application/json",
				"accept": "application/json",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).ToNot(BeEmpty())
			Expect(results[0]).To(ContainSubstring("duplicate case-variant headers"))
			Expect(results[0]).To(ContainSubstring("Accept"))
		})

		It("should fail validation when duplicate case-variant headers with different values", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Accept": "application/json",
				"accept": "text/plain",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).ToNot(BeEmpty())
			Expect(results[0]).To(ContainSubstring("duplicate case-variant headers"))
			Expect(results[0]).To(ContainSubstring("Accept"))
		})

		It("should fail validation with multiple duplicate case-variant headers", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Accept":       "application/json",
				"accept":       "application/json",
				"Content-Type": "text/plain",
				"content-type": "text/plain",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).ToNot(BeEmpty())
			// Should have errors for both invalid headers and duplicates
			Expect(len(results)).To(BeNumerically(">=", 2))
		})

		It("should fail validation with case-variant of forbidden header", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"authorization": "test",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).ToNot(BeEmpty())
			Expect(results[0]).To(ContainSubstring("invalid headers"))
		})

		It("should fail validation when mixing different case variants", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"X-Custom-Header": "value1",
				"x-custom-header": "value1",
				"X-CUSTOM-HEADER": "value1",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).ToNot(BeEmpty())
			Expect(results[0]).To(ContainSubstring("duplicate case-variant headers"))
			Expect(results[0]).To(ContainSubstring("X-Custom-Header"))
		})

		It("should produce deterministic messages for multiple forbidden headers", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Content-Type":  "text/plain",
				"Authorization": "test",
			}
			expected := validateElasticsearchHeaders(spec)
			for i := 0; i < 100; i++ {
				Expect(validateElasticsearchHeaders(spec)).To(Equal(expected),
					"validation messages should be stable across invocations")
			}
		})

		It("should produce deterministic messages for multiple duplicate header groups", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Accept":          "application/json",
				"accept":          "text/plain",
				"X-Custom-Header": "value1",
				"x-custom-header": "value2",
			}
			expected := validateElasticsearchHeaders(spec)
			for i := 0; i < 100; i++ {
				Expect(validateElasticsearchHeaders(spec)).To(Equal(expected),
					"validation messages should be stable across invocations")
			}
		})

		It("should produce sorted header names in forbidden headers message", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"Content-Type":  "text/plain",
				"Authorization": "test",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).To(HaveLen(1))
			Expect(results[0]).To(Equal("invalid headers found: Authorization,Content-Type"))
		})

		It("should produce sorted variant names in duplicate headers message", func() {
			spec.Elasticsearch.Headers = map[string]string{
				"X-Custom-Header": "value1",
				"x-custom-header": "value1",
				"X-CUSTOM-HEADER": "value1",
			}
			results := validateElasticsearchHeaders(spec)
			Expect(results).To(HaveLen(1))
			Expect(results[0]).To(Equal("duplicate case-variant headers 'X-CUSTOM-HEADER', 'X-Custom-Header', 'x-custom-header' found, use canonical form 'X-Custom-Header'"))
		})
	})
})

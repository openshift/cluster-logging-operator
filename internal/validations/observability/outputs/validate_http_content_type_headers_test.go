package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("[internal][validations] ClusterLogForwarder will validate Content-Type header in HTTP Output", func() {
	var (
		http *v1.HTTP
		spec v1.OutputSpec
	)
	BeforeEach(func() {
		http = &v1.HTTP{}
		spec = v1.OutputSpec{
			Name: "httpOutput",
			Type: v1.OutputTypeHTTP,
			HTTP: http,
		}
	})

	Context("#validateHttpContentTypeHeaders", func() {

		It("should pass validation with empty headers", func() {
			Expect(validateHttpContentTypeHeaders(spec)).To(BeEmpty())
		})
		It("should pass validation when not Content Type header", func() {
			spec.HTTP.Headers = map[string]string{
				"Accept": "application/json",
			}
			Expect(validateHttpContentTypeHeaders(spec)).To(BeEmpty())
		})
		It("should pass validation when the Content Type header is application/json", func() {
			spec.HTTP.Headers = map[string]string{
				"Content-Type": "application/json",
			}
			Expect(validateHttpContentTypeHeaders(spec)).To(BeEmpty())
		})
		It("should pass validation when the Content Type header is application/x-ndjson", func() {
			spec.HTTP.Headers = map[string]string{
				"Content-Type": "application/x-ndjson",
			}
			Expect(validateHttpContentTypeHeaders(spec)).To(BeEmpty())
		})
		It("should fail validation when not valid content types", func() {
			spec.HTTP.Headers = map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}
			Expect(validateHttpContentTypeHeaders(spec)).ToNot(BeEmpty())
		})
		It("should pass validation when not HTTP Output", func() {
			spec = v1.OutputSpec{
				Name:          "esOutput",
				Type:          v1.OutputTypeElasticsearch,
				Elasticsearch: &v1.Elasticsearch{},
			}
			Expect(validateHttpContentTypeHeaders(spec)).To(BeEmpty())
		})
	})
})

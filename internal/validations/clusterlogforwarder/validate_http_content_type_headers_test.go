package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("[internal][validations] ClusterLogForwarder will validate Content-Type header in Http Output", func() {
	var (
		clf  *v1.ClusterLogForwarder
		http *v1.Http
	)
	BeforeEach(func() {
		http = &v1.Http{}
		clf = &v1.ClusterLogForwarder{
			Spec: v1.ClusterLogForwarderSpec{
				Outputs: []v1.OutputSpec{
					{
						Name: "httpOutput",
						Type: v1.OutputTypeHttp,
						OutputTypeSpec: v1.OutputTypeSpec{
							Http: http,
						},
					},
				},
			},
		}
	})

	Context("#validateHttpContentTypeHeaders", func() {

		It("should pass validation with empty headers", func() {
			Expect(validateHttpContentTypeHeaders(*clf, nil, nil)).To(Succeed())
		})
		It("should pass validation when not Content Type header", func() {
			clf.Spec.Outputs[0].Http.Headers = map[string]string{
				"Accept": "application/json",
			}
			Expect(validateHttpContentTypeHeaders(*clf, nil, nil)).To(Succeed())
		})
		It("should pass validation when the Content Type header is application/json", func() {
			clf.Spec.Outputs[0].Http.Headers = map[string]string{
				"Content-Type": "application/json",
			}
			Expect(validateHttpContentTypeHeaders(*clf, nil, nil)).To(Succeed())
		})
		It("should pass validation when the Content Type header is application/x-ndjson", func() {
			clf.Spec.Outputs[0].Http.Headers = map[string]string{
				"Content-Type": "application/x-ndjson",
			}
			Expect(validateHttpContentTypeHeaders(*clf, nil, nil)).To(Succeed())
		})
		It("should fail validation when not valid content types", func() {
			clf.Spec.Outputs[0].Http.Headers = map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}
			Expect(validateHttpContentTypeHeaders(*clf, nil, nil)).To(Not(Succeed()))
		})
		It("should pass validation when not Http Output", func() {
			notHttpClf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Outputs: []v1.OutputSpec{
						{
							Name: "esOutput",
							Type: v1.OutputTypeElasticsearch,
							OutputTypeSpec: v1.OutputTypeSpec{
								Elasticsearch: &v1.Elasticsearch{},
							},
						},
					},
				},
			}
			clf.Spec.Outputs[0].Http.Headers = map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}
			Expect(validateHttpContentTypeHeaders(*notHttpClf, nil, nil)).To(Succeed())
		})
	})
})

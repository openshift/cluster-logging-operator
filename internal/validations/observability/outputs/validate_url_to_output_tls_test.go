package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("[internal][validations][observability][outputs] ClusterLogForwarder: Output URL vs Output TLS", func() {

	var spec obs.OutputSpec

	Context("#validateURLAccordingToTLS", func() {

		BeforeEach(func() {
			spec = obs.OutputSpec{
				Name: "myOutput",
			}
		})

		It("should fail validation when not secure URL and tls.InsecureSkipVerify=true", func() {
			spec.Type = obs.OutputTypeSplunk
			spec.Splunk = &obs.Splunk{
				URLSpec: obs.URLSpec{
					URL: "http://local.svc:514",
				},
			}
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
			}
			Expect(validateURLAccordingToTLS(spec)).To(Not(BeEmpty()))
		})
		It("should pass validation when not secure URL and no TLS config", func() {
			spec.Syslog = &obs.Syslog{
				URLSpec: obs.URLSpec{
					URL: "tcp://local.svc:514",
				},
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should pass validation when when not secure URL and tls.InsecureSkipVerify=false", func() {
			spec.Type = obs.OutputTypeSplunk
			spec.Splunk = &obs.Splunk{
				URLSpec: obs.URLSpec{
					URL: "http://local.svc:514",
				},
			}
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: false,
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should fail validation when not secure URL and exist TLS config: tls.TLSSecurityProfile", func() {
			spec.Type = obs.OutputTypeLoki
			spec.Loki = &obs.Loki{
				URLSpec: obs.URLSpec{
					URL: "http://local.svc:514",
				},
			}
			spec.TLS = &obs.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateURLAccordingToTLS(spec)).To(Not(BeEmpty()))
		})
		It("should pass validation when secure URL and exist TLS config: tls.InsecureSkipVerify=true", func() {
			spec.Type = obs.OutputTypeHTTP
			spec.HTTP = &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: "https://local.svc:514",
				},
			}
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should pass validation when secure URL and exist TLS config: tls.InsecureSkipVerify=false", func() {
			spec.Type = obs.OutputTypeKafka
			spec.Kafka = &obs.Kafka{
				URLSpec: obs.URLSpec{
					URL: "https://local.svc:514",
				},
			}
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: false,
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should pass validation when secure URL and exist TLS config: tls.TLSSecurityProfile", func() {
			spec.Type = obs.OutputTypeElasticsearch
			spec.Elasticsearch = &obs.Elasticsearch{
				URLSpec: obs.URLSpec{
					URL: "https://local.svc:514",
				},
			}

			spec.TLS = &obs.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should pass validation when URL is optional and TLS is spec'd", func() {
			spec.Type = obs.OutputTypeCloudwatch
			spec.Cloudwatch = &obs.Cloudwatch{}
			spec.TLS = &obs.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should pass validation when URL not provided for specific Output type", func() {
			spec.Type = obs.OutputTypeGoogleCloudLogging
			spec.GoogleCloudLogging = &obs.GoogleCloudLogging{}
			spec.TLS = &obs.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
	})
})

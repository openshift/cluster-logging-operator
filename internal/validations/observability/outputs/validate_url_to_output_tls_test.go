package outputs

import (
	. "github.com/onsi/ginkgo/v2"
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
				URL: "tcp://local.svc:514",
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
				URL: "https://local.svc:514",
			}
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: false,
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should pass validation when secure URL and exist TLS config: tls.TLSSecurityProfile", func() {
			spec.Type = obs.OutputTypeElasticsearch
			spec.Elasticsearch = &obs.Elasticsearch{
				URL: "https://local.svc:514",
			}

			spec.TLS = &obs.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateURLAccordingToTLS(spec)).To(BeEmpty())
		})
		It("should fail validation when any endpoint has insecure URL with TLS config", func() {
			spec.Type = obs.OutputTypeElasticsearch
			spec.Elasticsearch = &obs.Elasticsearch{
				URL: "https://secure.svc:9200",
				Endpoints: []obs.EndpointURL{
					"http://insecure.svc:9200",
				},
			}
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
			}
			Expect(validateURLAccordingToTLS(spec)).ToNot(BeEmpty())
		})
		It("should pass validation when all endpoints are secure with TLS config", func() {
			spec.Type = obs.OutputTypeElasticsearch
			spec.Elasticsearch = &obs.Elasticsearch{
				Endpoints: []obs.EndpointURL{
					"https://es1.svc:9200",
					"https://es2.svc:9200",
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

	Context("#validateHTTPBearerTokenRequiresHTTPS", func() {

		BeforeEach(func() {
			spec = obs.OutputSpec{
				Name: "myOutput",
				Type: obs.OutputTypeHTTP,
			}
		})

		It("should pass validation when HTTP bearer token not configured", func() {
			spec.HTTP = &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: "http://local.svc:8080",
				},
			}
			Expect(validateHTTPBearerTokenRequiresHTTPS(spec)).To(BeEmpty())
		})

		It("should pass validation when HTTP bearer token uses HTTPS URL", func() {
			spec.HTTP = &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: "https://secure.svc:8080",
				},
				Authentication: &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Name: "my-secret",
							Key:  "token",
						},
					},
				},
			}
			Expect(validateHTTPBearerTokenRequiresHTTPS(spec)).To(BeEmpty())
		})

		It("should fail validation when HTTP bearer token uses plain HTTP URL", func() {
			spec.HTTP = &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: "http://insecure.svc:8080",
				},
				Authentication: &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Name: "my-secret",
							Key:  "token",
						},
					},
				},
			}
			results := validateHTTPBearerTokenRequiresHTTPS(spec)
			Expect(results).ToNot(BeEmpty())
			Expect(results[0]).To(ContainSubstring("HTTP bearer token authentication requires HTTPS URL"))
		})

		It("should fail validation when HTTP bearer token from serviceAccount uses plain HTTP URL", func() {
			spec.HTTP = &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: "http://insecure.svc:8080",
				},
				Authentication: &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromServiceAccount,
					},
				},
			}
			results := validateHTTPBearerTokenRequiresHTTPS(spec)
			Expect(results).ToNot(BeEmpty())
			Expect(results[0]).To(ContainSubstring("HTTP bearer token authentication requires HTTPS URL"))
		})

		It("should pass validation when HTTP bearer token is not set even with HTTPS", func() {
			spec.HTTP = &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: "https://secure.svc:8080",
				},
				Authentication: &obs.HTTPAuthentication{
					Username: &obs.SecretReference{
						SecretName: "my-secret",
						Key:        "username",
					},
				},
			}
			Expect(validateHTTPBearerTokenRequiresHTTPS(spec)).To(BeEmpty())
		})

		It("should pass validation for non-HTTP output types", func() {
			spec.Type = obs.OutputTypeLoki
			spec.Loki = &obs.Loki{
				URLSpec: obs.URLSpec{
					URL: "http://insecure.svc:3100",
				},
				Authentication: &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Name: "my-secret",
							Key:  "token",
						},
					},
				},
			}
			Expect(validateHTTPBearerTokenRequiresHTTPS(spec)).To(BeEmpty())
		})
	})
})

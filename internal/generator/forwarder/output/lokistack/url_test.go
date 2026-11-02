package lokistack_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/output/lokistack"
)

var _ = Describe("LokiStack URL utilities", func() {
	Context("GatewayService", func() {
		It("should return the correct gateway service name", func() {
			service := lokistack.GatewayService("logging-loki")
			Expect(service).To(Equal("logging-loki-gateway-http"))
		})
	})

	Context("URL", func() {
		It("should build base URL for application tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "logging-loki",
					Namespace: "openshift-logging",
				},
			}

			url := lokistack.URL(ls, "application")
			Expect(url).To(Equal("https://logging-loki-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"))
		})

		It("should build base URL for infrastructure tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "my-loki",
					Namespace: "my-namespace",
				},
			}

			url := lokistack.URL(ls, "infrastructure")
			Expect(url).To(Equal("https://my-loki-gateway-http.my-namespace.svc:8080/api/logs/v1/infrastructure"))
		})

		It("should build base URL for audit tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "logging-loki",
					Namespace: "openshift-logging",
				},
			}

			url := lokistack.URL(ls, "audit")
			Expect(url).To(Equal("https://logging-loki-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit"))
		})

		It("should return empty string for invalid tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "logging-loki",
					Namespace: "openshift-logging",
				},
			}

			url := lokistack.URL(ls, "invalid")
			Expect(url).To(BeEmpty())
		})
	})

	Context("OtlpURL", func() {
		It("should build OTLP URL for application tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "logging-loki",
					Namespace: "openshift-logging",
				},
			}

			url := lokistack.OtlpURL(ls, "application")
			Expect(url).To(Equal("https://logging-loki-gateway-http.openshift-logging.svc:8080/api/logs/v1/application/otlp/v1/logs"))
		})

		It("should build OTLP URL for infrastructure tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "my-loki",
					Namespace: "my-namespace",
				},
			}

			url := lokistack.OtlpURL(ls, "infrastructure")
			Expect(url).To(Equal("https://my-loki-gateway-http.my-namespace.svc:8080/api/logs/v1/infrastructure/otlp/v1/logs"))
		})

		It("should build OTLP URL for audit tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "logging-loki",
					Namespace: "openshift-logging",
				},
			}

			url := lokistack.OtlpURL(ls, "audit")
			Expect(url).To(Equal("https://logging-loki-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit/otlp/v1/logs"))
		})

		It("should return empty string for invalid tenant", func() {
			ls := &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "logging-loki",
					Namespace: "openshift-logging",
				},
			}

			url := lokistack.OtlpURL(ls, "invalid")
			Expect(url).To(BeEmpty())
		})
	})
})

package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Exporters", func() {
	Context("OtlpHttp Exporter", func() {
		It("should marshal to YAML correctly", func() {
			exporter := exporters.NewOtlpHttp("", "http://loki:3100/otlp")
			exporter.Encoding = "proto"
			exporter.Compression = "gzip"
			exporter.Timeout = "30s"
			exporter.Headers = map[string]string{
				"X-Scope-OrgID": "tenant1",
			}
			exporter.TLS = &types.TlsClientConfig{
				Insecure: true,
			}

			data, err := yaml.Marshal(exporter)
			Expect(err).To(BeNil())
			Expect(string(data)).To(ContainSubstring("endpoint: http://loki:3100/otlp"))
			Expect(string(data)).To(ContainSubstring("encoding: proto"))
			Expect(string(data)).To(ContainSubstring("compression: gzip"))
			Expect(string(data)).To(ContainSubstring("timeout: 30s"))
			Expect(string(data)).To(ContainSubstring("X-Scope-OrgID: tenant1"))
		})

		It("should unmarshal from YAML correctly", func() {
			yamlData := `
endpoint: http://loki:3100/otlp
encoding: json
compression: gzip
timeout: 60s
headers:
  X-Scope-OrgID: production
  X-Custom-Header: custom-value
tls:
  insecure: true
sending_queue:
  enabled: true
  num_consumers: 5
  queue_size: 500
retry_on_failure:
  enabled: true
  initial_interval: 5s
  max_interval: 30s
  max_elapsed_time: 5m
`
			var exporter exporters.OtlpHttp
			err := yaml.Unmarshal([]byte(yamlData), &exporter)
			Expect(err).To(BeNil())
			Expect(exporter.Endpoint).To(Equal("http://loki:3100/otlp"))
			Expect(exporter.Encoding).To(Equal("json"))
			Expect(exporter.Compression).To(Equal("gzip"))
			Expect(exporter.Timeout).To(Equal("60s"))
			Expect(exporter.Headers).To(HaveKeyWithValue("X-Scope-OrgID", "production"))
			Expect(exporter.Headers).To(HaveKeyWithValue("X-Custom-Header", "custom-value"))
			Expect(exporter.TLS).ToNot(BeNil())
			Expect(exporter.TLS.Insecure).To(BeTrue())
			Expect(exporter.SendingQueue).ToNot(BeNil())
			Expect(exporter.SendingQueue.Enabled).To(BeTrue())
			Expect(exporter.SendingQueue.NumConsumers).To(Equal(5))
			Expect(exporter.RetryOnFailure).ToNot(BeNil())
			Expect(exporter.RetryOnFailure.Enabled).To(BeTrue())
		})
	})

	Context("Exporters Map", func() {
		It("should unmarshal multiple exporters from YAML", func() {
			yamlData := `
otlphttp/loki:
  endpoint: http://loki:3100/otlp
  encoding: proto
  headers:
    X-Scope-OrgID: default
otlphttp/production:
  endpoint: http://loki-prod:3100/otlp
  encoding: json
  timeout: 60s
  tls:
    insecure: false
    ca_file: /etc/certs/ca.pem
`
			var exportersMap api.Exporters
			err := yaml.Unmarshal([]byte(yamlData), &exportersMap)
			Expect(err).To(BeNil())
			Expect(exportersMap).To(HaveLen(2))

			lokiExporter, ok := exportersMap["otlphttp/loki"]
			Expect(ok).To(BeTrue())
			otlpLoki, ok := lokiExporter.(*exporters.OtlpHttp)
			Expect(ok).To(BeTrue())
			Expect(otlpLoki.Endpoint).To(Equal("http://loki:3100/otlp"))
			Expect(otlpLoki.Encoding).To(Equal("proto"))
			Expect(otlpLoki.Headers).To(HaveKeyWithValue("X-Scope-OrgID", "default"))

			prodExporter, ok := exportersMap["otlphttp/production"]
			Expect(ok).To(BeTrue())
			otlpProd, ok := prodExporter.(*exporters.OtlpHttp)
			Expect(ok).To(BeTrue())
			Expect(otlpProd.Endpoint).To(Equal("http://loki-prod:3100/otlp"))
			Expect(otlpProd.Encoding).To(Equal("json"))
			Expect(otlpProd.Timeout).To(Equal("60s"))
			Expect(otlpProd.TLS).ToNot(BeNil())
			Expect(otlpProd.TLS.Insecure).To(BeFalse())
			Expect(otlpProd.TLS.CAFile).To(Equal("/etc/certs/ca.pem"))
		})
	})
})

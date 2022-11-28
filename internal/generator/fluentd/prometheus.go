package fluentd

type PrometheusMonitor struct {
	TlsMinVersion string
	CipherSuites  string
}

func (p PrometheusMonitor) Name() string {
	return "PrometheusMonitor"
}

func (p PrometheusMonitor) Template() string {
	return `{{define "` + p.Name() + `" -}}
  # Prometheus Monitoring
  <source>
    @type prometheus
    bind "#{ENV['PROM_BIND_IP']}"
    <transport tls>
      cert_path /etc/collector/metrics/tls.crt
      private_key_path /etc/collector/metrics/tls.key
      min_version {{.TlsMinVersion}}
      max_version TLS1_3
      ciphers {{.CipherSuites}}
    </transport>
  </source>

  <source>
    @type prometheus_monitor
    <labels>
      hostname ${hostname}
    </labels>
  </source>

  # excluding prometheus_tail_monitor
  # since it leaks namespace/pod info
  # via file paths

  # tail_monitor plugin which publishes log_collected_bytes_total
  <source>
    @type collected_tail_monitor
    <labels>
      hostname ${hostname}
    </labels>
  </source>

  # This is considered experimental by the repo
  <source>
    @type prometheus_output_monitor
    <labels>
      hostname ${hostname}
    </labels>
  </source>
  {{end}}`
}

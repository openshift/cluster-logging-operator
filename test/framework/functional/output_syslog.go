package functional

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"net/url"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"

	log "github.com/ViaQ/logerr/v2/log/static"
)

const IncreaseRsyslogMaxMessageSize = "$MaxMessageSize 50000"

func (f *CollectorFunctionalFramework) AddSyslogOutput(b *runtime.PodBuilder, output obs.OutputSpec) error {
	log.V(2).Info("Adding syslog output", "name", output.Name)
	name := strings.ToLower(output.Name)
	var baseRsyslogConfig string
	u, _ := url.Parse(output.Syslog.URL)
	if strings.ToLower(u.Scheme) == "udp" {
		baseRsyslogConfig = e2e.UdpSyslogInput
		if output.TLS != nil {
			baseRsyslogConfig = e2e.UdpSyslogInputWithTLS
		}
	} else {
		baseRsyslogConfig = e2e.TcpSyslogInput
		if output.TLS != nil {
			baseRsyslogConfig = e2e.TcpSyslogInputWithTLS
		}
	}

	// using unsecure rsyslog conf
	rfc := e2e.RFC5424
	if output.Syslog != nil && output.Syslog.RFC != "" {
		rfc = e2e.MustParseRFC(string(output.Syslog.RFC))
	}
	rsyslogConf := e2e.GenerateRsyslogConf(baseRsyslogConfig, rfc)
	rsyslogConf = strings.Join([]string{IncreaseRsyslogMaxMessageSize, rsyslogConf}, "\n")
	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "rsyslog.conf", rsyslogConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding container", "name", name)
	containerBuilder := b.AddContainer(name, e2e.ImageRemoteSyslog).
		AddVolumeMount(config.Name, "/rsyslog/etc", "", false).
		WithCmdArgs([]string{"rsyslogd", "-n", "-f", "/rsyslog/etc/rsyslog.conf"}).
		WithPrivilege()
	if output.TLS != nil {
		for _, name := range internalobs.SecretsForTLS(output.TLS.TLSSpec) {
			containerBuilder.AddVolumeMount(name, "/rsyslog/etc/secrets", "", true)
		}
	}
	containerBuilder.End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

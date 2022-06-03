package functional

import (
	"net/url"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"

	"github.com/ViaQ/logerr/v2/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const ImageRemoteSyslog = "quay.io/openshift/origin-logging-rsyslog:latest"

const IncreaseRsyslogMaxMessageSize = "$MaxMessageSize 50000"

func (f *CollectorFunctionalFramework) addSyslogOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	logger := log.NewLogger("output-syslog-testing")

	logger.V(2).Info("Adding syslog output", "name", output.Name)
	name := strings.ToLower(output.Name)
	var baseRsyslogConfig string
	u, _ := url.Parse(output.URL)
	if strings.ToLower(u.Scheme) == "udp" {
		baseRsyslogConfig = e2e.UdpSyslogInput
	} else {
		baseRsyslogConfig = e2e.TcpSyslogInput
	}
	// using unsecure rsyslog conf
	rfc := e2e.RFC5424
	if output.Syslog != nil && output.Syslog.RFC != "" {
		rfc = e2e.MustParseRFC(output.Syslog.RFC)
	}
	rsyslogConf := e2e.GenerateRsyslogConf(baseRsyslogConfig, rfc)
	rsyslogConf = strings.Join([]string{IncreaseRsyslogMaxMessageSize, rsyslogConf}, "\n")
	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	logger.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "rsyslog.conf", rsyslogConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	logger.V(2).Info("Adding container", "name", name)
	b.AddContainer(name, ImageRemoteSyslog).
		AddVolumeMount(config.Name, "/rsyslog/etc", "", false).
		WithCmdArgs([]string{"rsyslogd", "-n", "-f", "/rsyslog/etc/rsyslog.conf"}).
		WithPrivilege().
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

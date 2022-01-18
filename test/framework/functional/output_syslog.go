package functional

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"net/url"
	"strings"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const ImageRemoteSyslog = "quay.io/openshift/origin-logging-rsyslog:latest"

const IncreaseRsyslogMaxMessageSize = "$MaxMessageSize 50000"

func (f *CollectorFunctionalFramework) addSyslogOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding syslog output", "name", output.Name)
	name := strings.ToLower(output.Name)
	var baseRsyslogConfig string
	u, _ := url.Parse(output.URL)
	if strings.ToLower(u.Scheme) == "udp" {
		baseRsyslogConfig = e2e.UdpSyslogInput
	} else {
		baseRsyslogConfig = e2e.TcpSyslogInput
	}
	// using unsecure rsyslog conf
	rsyslogConf := e2e.GenerateRsyslogConf(baseRsyslogConfig, e2e.RFC5424)
	rsyslogConf = strings.Join([]string{IncreaseRsyslogMaxMessageSize, rsyslogConf}, "\n")
	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "rsyslog.conf", rsyslogConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding container", "name", name)
	b.AddContainer(name, ImageRemoteSyslog).
		AddVolumeMount(config.Name, "/rsyslog/etc", "", false).
		WithCmdArgs([]string{"rsyslogd", "-n", "-f", "/rsyslog/etc/rsyslog.conf"}).
		WithPrivilege().
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

package fluentd

func (conf *outputLabelConf) SyslogLegacyPlugin() string {
	if protocol := conf.Protocol(); protocol == "udp" {
		return "syslog"
	}
	return "syslog_buffered"
}

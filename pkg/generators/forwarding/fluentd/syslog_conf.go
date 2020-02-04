package fluentd

func (conf *outputLabelConf) SyslogLegacyPlugin() string {
	if protocol := conf.Protocol(); protocol == "udp" {
		return "syslog"
	}
	return "syslog_buffered"
}

func (conf *outputLabelConf) SyslogProtocol() string {
	if protocol := conf.Protocol(); protocol != "" {
		return protocol
	}
	return "tcp"
}

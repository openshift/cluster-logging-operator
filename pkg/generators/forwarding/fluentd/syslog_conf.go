package fluentd

func (conf *outputLabelConf) SyslogPlugin() string {
	if protocol := conf.Protocol(); protocol == "udp" {
		return "syslog"
	}
	return "syslog_buffered"
}

package fluentd

func (conf *outputLabelConf) SyslogPlugin() string {
	if protocol := conf.Protocol(); protocol == "udp" {
		return "syslog"
	}
	return "syslog_buffered"
}

func (conf *outputLabelConf) Rfc() string {
	//return "rfc3164"
	// TODO update it whem merged with new v1 api
	return "rfc5424"
}

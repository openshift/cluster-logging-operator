package e2e

const (
	TcpSyslogInput = `
# Provides TCP syslog reception
# for parameters see http://www.rsyslog.com/doc/imtcp.html
module(load="imtcp") # needs to be done just once
input(type="imtcp" port="24224" ruleset="test")
	`

	TcpSyslogInputWithTLS = `
# Provides TCP syslog reception
# for parameters see http://www.rsyslog.com/doc/imtcp.html
module(load="imtcp"
    StreamDriver.Name="gtls"
    StreamDriver.Mode="1" # run driver in TLS-only mode
    StreamDriver.Authmode="x509/certvalid"
)
# make gtls driver the default and set certificate files
global(
    DefaultNetstreamDriver="gtls"
    DefaultNetstreamDriverCAFile="/rsyslog/etc/secrets/ca-bundle.crt"
    DefaultNetstreamDriverCertFile="/rsyslog/etc/secrets/tls.crt"
    DefaultNetstreamDriverKeyFile="/rsyslog/etc/secrets/tls.key"
    )
input(type="imtcp" port="24224" ruleset="test")
	`

	UdpSyslogInput = `
# Provides UDP syslog reception
# for parameters see http://www.rsyslog.com/doc/imudp.html
module(load="imudp") # needs to be done just once
input(type="imudp" port="24224" ruleset="test")
	`

	UdpSyslogInputWithTLS = `
# Provides UDP syslog reception
# for parameters see http://www.rsyslog.com/doc/imudp.html
module(load="imudp"
    StreamDriver.Name="gtls"
    StreamDriver.Mode="1" # run driver in TLS-only mode
    StreamDriver.Authmode="x509/certvalid"
) # needs to be done just once
# make gtls driver the default and set certificate files
global(
    DefaultNetstreamDriver="gtls"
    DefaultNetstreamDriverCAFile="/rsyslog/etc/secrets/ca-bundle.crt"
    DefaultNetstreamDriverCertFile="/rsyslog/etc/secrets/tls.crt"
    DefaultNetstreamDriverKeyFile="/rsyslog/etc/secrets/tls.key"
    )
input(type="imudp" port="24224" ruleset="test")
	`

	RuleSetRfc5424 = `
#### RULES ####
ruleset(name="test" parser=["rsyslog.rfc5424"]){
    action(type="omfile" file="/tmp/infra.log" Template="RSYSLOG_SyslogProtocol23Format")
}
	`

	RuleSetRfc3164 = `
#### RULES ####
ruleset(name="test" parser=["rsyslog.rfc3164"]){
    action(type="omfile" file="/tmp/infra.log" Template="RSYSLOG_SyslogProtocol23Format")
}
	`

	RuleSetRfc3164WithPRI = `
# Define the RFC3164 template including <PRI>
template(name="RFC3164WithPRI" type="string"
         string="<%PRI%>%TIMESTAMP% %HOSTNAME% %syslogtag%%msg%\n")

#### RULES ####
ruleset(name="test" parser=["rsyslog.rfc3164"]) {
    action(type="omfile" file="/tmp/infra.log" Template="RFC3164WithPRI")
}
    `
	// includes both rfc parsers
	RuleSetRfc3164Rfc5424 = `
#### RULES ####
ruleset(name="test" parser=["rsyslog.rfc3164","rsyslog.rfc5424"]){
    action(type="omfile" file="/tmp/infra.log" Template="RSYSLOG_SyslogProtocol23Format")
}
	`
)

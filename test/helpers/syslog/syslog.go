package syslog

import (
	"log"
	"strings"
)

const (
	ImageRemoteSyslog = "registry.redhat.io/rhel9/rsyslog:9.8-1780596788"
)

// SyslogRfc type is the rfc used for sending syslog
type SyslogRfc int

const (
	// RFC3164 rfc3164
	RFC3164 SyslogRfc = iota
	// RFC5424 rfc5424
	RFC5424
	// RFC3164RFC5424 either rfc3164 or rfc5424
	RFC3164RFC5424
)

func MustParseRFC(rfc string) SyslogRfc {
	switch strings.ToUpper(rfc) {
	case "RFC3164":
		return RFC3164
	case "RFC5424":
		return RFC5424
	case "RFC3164 OR RFC5424":
		return RFC3164RFC5424
	}
	log.Fatal("Unable to parse RFC", "rfc", rfc)
	return 0
}

func (e SyslogRfc) String() string {
	switch e {
	case RFC3164:
		return "RFC3164"
	case RFC5424:
		return "RFC5424"
	case RFC3164RFC5424:
		return "RFC3164 or RFC5424"
	default:
		return "Unknown rfc"
	}
}

func GenerateRsyslogConf(conf string, rfc SyslogRfc) string {
	switch rfc {
	case RFC5424:
		return strings.Join([]string{conf, RuleSetRfc5424}, "\n")
	case RFC3164:
		return strings.Join([]string{conf, RuleSetRfc3164}, "\n")
	case RFC3164RFC5424:
		return strings.Join([]string{conf, RuleSetRfc3164Rfc5424}, "\n")
	}
	return "Invalid Conf"
}

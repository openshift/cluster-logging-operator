package syslog

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapSyslog(loggingOutSpec logging.OutputSpec) *obs.Syslog {
	obsSyslog := &obs.Syslog{
		URL:      loggingOutSpec.URL,
		RFC:      obs.SyslogRFC5424,
		Facility: "user",
		Severity: "informational",
	}

	loggingSyslog := loggingOutSpec.Syslog
	if loggingSyslog == nil {
		return obsSyslog
	}

	if loggingSyslog.RFC != "" {
		obsSyslog.RFC = obs.SyslogRFCType(loggingSyslog.RFC)
	}

	if loggingSyslog.Facility != "" {
		obsSyslog.Facility = loggingSyslog.Facility
	}

	if loggingSyslog.Severity != "" {
		obsSyslog.Severity = loggingSyslog.Severity
	}

	if loggingSyslog.AddLogSource {
		obsSyslog.Enrichment = obs.EnrichmentTypeKubernetesMinimal
	}

	obsSyslog.AppName = loggingSyslog.AppName
	if strings.HasPrefix(loggingSyslog.AppName, "$.message") {
		obsSyslog.AppName = fmt.Sprintf(`{%s||"none"}`, strings.Split(loggingSyslog.AppName, "$.message")[1])
	}

	obsSyslog.ProcID = loggingSyslog.ProcID
	if strings.HasPrefix(loggingSyslog.ProcID, "$.message") {
		obsSyslog.AppName = fmt.Sprintf(`{%s||"none"}`, strings.Split(loggingSyslog.ProcID, "$.message")[1])
	}

	obsSyslog.MsgID = loggingSyslog.MsgID
	if strings.HasPrefix(loggingSyslog.MsgID, "$.message") {
		obsSyslog.AppName = fmt.Sprintf(`{%s||"none"}`, strings.Split(loggingSyslog.MsgID, "$.message")[1])
	}

	if loggingSyslog.PayloadKey != "" {
		obsSyslog.PayloadKey = fmt.Sprintf("{.%s}", loggingSyslog.PayloadKey)
	}

	return obsSyslog
}

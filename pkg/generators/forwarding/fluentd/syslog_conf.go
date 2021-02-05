package fluentd

import (
	"fmt"
	"regexp"
	"strings"
)

// The Syslog output fields can be set to an expression of the form $.abc.xyz
// If an expression is used, its value will be taken from corresponding key in the record
var keyre = regexp.MustCompile(`^\$(\.[[:word:]]*)+$`)

var tagre = regexp.MustCompile(`\${tag\[-??\d+\]}`)

func (conf *outputLabelConf) SyslogPlugin() string {
	if protocol := conf.Protocol(); protocol == "udp" {
		return "syslog"
	}
	return "syslog_buffered"
}

func (conf *outputLabelConf) Facility() string {
	facility := strings.ToLower(conf.Target.Syslog.Facility)
	if facility == "" {
		return "user"
	}
	if conf.isFacilityKeyExpr() {
		return fmt.Sprintf("${%s}", facility)
	}
	return conf.Target.Syslog.Facility
}

func (conf *outputLabelConf) Severity() string {
	severity := strings.ToLower(conf.Target.Syslog.Severity)
	if severity == "" {
		return "debug"
	}
	if conf.isSeverityKeyExpr() {
		return fmt.Sprintf("${%s}", severity)
	}
	return severity
}

func (conf *outputLabelConf) Rfc() string {
	switch strings.ToLower(conf.Target.Syslog.RFC) {
	case "rfc3164":
		return "rfc3164"
	case "rfc5424":
		return "rfc5424"
	}
	return "Unknown Rfc"
}

func (conf *outputLabelConf) TrimPrefix() string {
	return conf.Target.Syslog.TrimPrefix
}

func (conf *outputLabelConf) Tag() string {
	tag := conf.Target.Syslog.Tag
	if tag == "" {
		return "-"
	}
	if conf.isTagKeyExpr() {
		return fmt.Sprintf("${%s}", tag)
	}
	if conf.isTagTagExpr() {
		return tag
	}
	if tag == "tag" {
		return "${tag}"
	}
	return tag
}

func (conf *outputLabelConf) PayloadKey() string {
	return conf.Target.Syslog.PayloadKey
}

func (conf *outputLabelConf) AddLogSource() bool {
	return conf.Target.Syslog.AddLogSource
}

func (conf *outputLabelConf) AppName() string {
	appname := conf.Target.Syslog.AppName
	if appname == "" {
		return "-"
	}
	if conf.isAppNameKeyExpr() {
		return fmt.Sprintf("${%s}", appname)
	}
	if conf.isTagTagExpr() {
		return appname
	}
	if appname == "tag" {
		return "${tag}"
	}
	return appname
}

func (conf *outputLabelConf) MsgID() string {
	msgid := conf.Target.Syslog.MsgID
	if msgid == "" {
		return "-"
	}
	if conf.isMsgIDKeyExpr() {
		return fmt.Sprintf("${%s}", msgid)
	}
	return msgid
}

func (conf *outputLabelConf) ProcID() string {
	procid := conf.Target.Syslog.ProcID
	if procid == "" {
		return "-"
	}
	if conf.isProcIDKeyExpr() {
		return fmt.Sprintf("${%s}", procid)
	}
	return procid
}

func (conf *outputLabelConf) ChunkKeys() string {
	keys := []string{}
	tagAdded := false
	if conf.isTagKeyExpr() {
		keys = append(keys, conf.Target.Syslog.Tag)
	}
	if conf.isTagTagExpr() && !tagAdded {
		keys = append(keys, "tag")
		tagAdded = true
	}
	if conf.Target.Syslog.Tag == "tag" && !tagAdded {
		keys = append(keys, "tag")
		tagAdded = true
	}
	if conf.isAppNameKeyExpr() {
		keys = append(keys, conf.Target.Syslog.AppName)
	}
	if conf.isAppNameTagExpr() && !tagAdded {
		keys = append(keys, "tag")
		tagAdded = true
	}
	if conf.Target.Syslog.AppName == "tag" && !tagAdded {
		keys = append(keys, "tag")
	}
	if conf.isMsgIDKeyExpr() {
		keys = append(keys, conf.Target.Syslog.MsgID)
	}
	if conf.isProcIDKeyExpr() {
		keys = append(keys, conf.Target.Syslog.ProcID)
	}
	if conf.isFacilityKeyExpr() {
		keys = append(keys, conf.Target.Syslog.Facility)
	}
	if conf.isSeverityKeyExpr() {
		keys = append(keys, conf.Target.Syslog.Severity)
	}
	return strings.Join(keys, ",")
}

func (conf *outputLabelConf) IsKeyExpr(str string) bool {
	return keyre.MatchString(str)
}

func (conf *outputLabelConf) IsTagExpr(str string) bool {
	return tagre.MatchString(str)
}

func (conf *outputLabelConf) isFacilityKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.Facility)
}

func (conf *outputLabelConf) isSeverityKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.Severity)
}

func (conf *outputLabelConf) isTagKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.Tag)
}

func (conf *outputLabelConf) isTagTagExpr() bool {
	return conf.IsTagExpr(conf.Target.Syslog.Tag)
}

func (conf *outputLabelConf) IsPayloadKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.PayloadKey)
}

func (conf *outputLabelConf) isAppNameKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.AppName)
}

func (conf *outputLabelConf) isAppNameTagExpr() bool {
	return conf.IsTagExpr(conf.Target.Syslog.AppName)
}

func (conf *outputLabelConf) isMsgIDKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.MsgID)
}

func (conf *outputLabelConf) isProcIDKeyExpr() bool {
	return conf.IsKeyExpr(conf.Target.Syslog.ProcID)
}

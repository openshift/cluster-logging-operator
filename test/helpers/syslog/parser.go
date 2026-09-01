package syslog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SyslogMessage struct {
	Priority       int
	Version        int
	Timestamp      time.Time
	Hostname       string
	AppName        string
	ProcID         string
	MsgID          string
	StructuredData string
	MessagePayload string
}

// ParseRFC5424SyslogLogs takes raw output and returns structured SyslogMessage objects.
func ParseRFC5424SyslogLogs(rawLog string) (*SyslogMessage, error) {
	if rawLog == "" {
		return &SyslogMessage{}, nil
	}
	rawLog = strings.TrimSpace(rawLog)
	syslogRegex := regexp.MustCompile(
		`^<(\d{1,3})>1\s` + //PRI VERSION
			`([^\s]+)\s` + // TIMESTAMP
			`([^\s]+)\s` + // HOSTNAME
			`([^\s]+)\s` + // APP-NAME
			`([^\s]+)\s` + // PROCID
			`([^\s]+)\s` + // MSGID
			`(-|\[[^\]]+\])\s` + // STRUCTURED-DATA
			`(.*)$`) // MESSAGE-PAYLOAD

	matches := syslogRegex.FindStringSubmatch(rawLog)

	if len(matches) != 9 {
		return nil, fmt.Errorf("failed to parse Syslog line: expected 9 submatches (got %d) in line: %s", len(matches), rawLog)
	}
	priority, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}
	ts, err := time.Parse(time.RFC3339Nano, matches[2])
	if err != nil {
		return nil, err
	}
	msg := &SyslogMessage{
		Priority:       priority,
		Version:        1,
		Timestamp:      ts,
		Hostname:       matches[3],
		AppName:        matches[4],
		ProcID:         matches[5],
		MsgID:          matches[6],
		StructuredData: matches[7],
		MessagePayload: matches[8],
	}
	return msg, nil
}

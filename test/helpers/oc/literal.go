package oc

import (
	"fmt"
	"strings"
)

// For oc commands not a part of this package, e.g. oc.Logs, oc.Apply, oc.Delete etc,
// oc.Literal is a workaround to run those commands.

// ILiteral is an interface for collecting the command string
type ILiteral interface {
	Command

	// an oc command string
	From(string, ...interface{}) ILiteral
}

type literal struct {
	*runner

	cmdstr string
}

// Literal takes an oc command string and runs it
func Literal() ILiteral {
	l := &literal{
		runner: &runner{},
	}
	l.collectArgsFunc = l.args
	return l
}

func (l *literal) WithConfig(cfg string) ILiteral {
	l.configPath = cfg
	return l
}

func (l *literal) From(cmd string, args ...interface{}) ILiteral {
	l.cmdstr = fmt.Sprintf(strings.TrimSpace(cmd), args...)
	return l
}

func (l *literal) String() string {
	split := strings.SplitN(l.cmdstr, " ", 2)
	if len(split) != 2 {
		return "command too small"
	}
	if split[0] != CMD {
		return "error: command string must start with 'oc'"
	}
	return sanitizeArgStr(fmt.Sprintf("%s %s", l.runner.String(), split[1]))
}

func (l *literal) args() []string {
	split := strings.SplitN(l.cmdstr, " ", 2)
	if len(split) != 2 {
		return []string{"--help"}
	}
	if split[0] != CMD {
		return []string{"--help"}
	}
	return sanitizeArgs(split[1])
}

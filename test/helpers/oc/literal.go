// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

	// an os command string
	From(string) ILiteral
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

func (l *literal) From(cmd string) ILiteral {
	l.cmdstr = strings.TrimSpace(cmd)
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

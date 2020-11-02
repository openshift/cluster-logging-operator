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
	"bytes"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"
)

// Runner is for executing the command. It provides implementation for
// the methods in oc.Command interface.
// Other commands like oc.Exec, oc.Get, oc.Literal collect their arguments
// and use Runner to run the commad with arguments.
// It provides different modes of executing the commands, Run/RunFor/Output/OutputFor
//
// As fas as possible, it is to be kept independent of oc command syntax.
// TODO(vimalk78)
// Move KUBECONFIG out from here

// CMD is the command to be run by the runner
const CMD string = "oc"

// runner encapsulates os/exec/Cmd, collects args, and runs CMD
type runner struct {
	*osexec.Cmd

	args       []string
	configPath string

	// This must be set by oc.Commands to collect arguments before calling Run
	collectArgsFunc func() []string

	tostdout bool

	err error
}

func (r *runner) Run() (string, error) {
	if r.err != nil {
		return "composed command failed", r.err
	}
	r.setArgs(r.collectArgsFunc())
	return r.runCmd()
}

func (r *runner) runCmd() (string, error) {
	// #nosec G204
	r.Cmd = osexec.Command(CMD, r.args...)
	var outbuf bytes.Buffer
	var errbuf bytes.Buffer
	if r.tostdout {
		r.Cmd.Stdout = os.Stdout
		r.Cmd.Stderr = os.Stderr
	} else {
		r.Cmd.Stdout = &outbuf
		r.Cmd.Stderr = &errbuf
	}
	r.Cmd.Env = []string{fmt.Sprintf("%s=%s", "KUBECONFIG", os.Getenv("KUBECONFIG"))}
	log.Info("running: ", "command", r, "arguments", strings.Join(r.args, " "))
	err := r.Cmd.Run()
	if err != nil {
		if r.tostdout {
			return "", err
		}
		errout := strings.TrimSpace(errbuf.String())
		log.Info("output", errout, "error", err)
		return errout, err
	}
	if r.tostdout {
		return "", nil
	}
	out := strings.TrimSpace(outbuf.String())
	if len(out) > 500 {
		log.Info("output(truncated 500/length)", "length", len(out), "result", truncateString(out, 500))
	} else {
		log.Info("output", out)
	}
	return out, nil
}

func (r *runner) RunFor(d time.Duration) (string, error) {
	time.AfterFunc(d, func() {
		_ = r.Kill()
	})
	return r.Run()
}

func (r *runner) Kill() error {
	if r.Process != nil {
		return r.Process.Kill()
	}
	return nil
}

func (r *runner) Output() error {
	r.tostdout = true
	_, err := r.Run()
	return err
}

func (r *runner) OutputFor(d time.Duration) error {
	r.tostdout = true
	_, err := r.RunFor(d)
	return err
}

func (r *runner) String() string {
	if r.configPath != "" {
		return fmt.Sprintf("%s --config %s", CMD, r.configPath)
	}
	return CMD
}

func sanitizeArgStr(argstr string) string {
	return strings.Join(sanitizeArgs(argstr), " ")
}

// sanitize the args, removes any unwanted spaces
func sanitizeArgs(argstr string) []string {
	outargs := []string{}
	args := strings.Split(argstr, " ")
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			outargs = append(outargs, arg)
		}
	}
	return outargs
}

func (r *runner) setArgs(args []string) {
	r.args = args
}

func truncateString(str string, num int) string {
	trunc := str
	if len(str) > num {
		if num > 4 {
			num -= 4
		}
		trunc = str[0:num] + " ..."
	}
	return trunc
}

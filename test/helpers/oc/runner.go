package oc

import (
	"bytes"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
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
	logger.Infof("running: %s %s", r, strings.Join(r.args, " "))
	err := r.Cmd.Run()
	if err != nil {
		if r.tostdout {
			return "", err
		}
		errout := strings.TrimSpace(errbuf.String())
		logger.Infof("output: %s, error: %v", errout, err)
		return errout, err
	}
	if r.tostdout {
		return "", nil
	}
	out := strings.TrimSpace(outbuf.String())
	logger.Infof("output: %s", out)
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

func (r *runner) setArgsStr(argstr string) {
	r.args = sanitizeArgs(argstr)
}

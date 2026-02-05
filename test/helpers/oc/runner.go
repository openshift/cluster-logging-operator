package oc

import (
	"bytes"
	"fmt"
	"os"
	osexec "os/exec"
	"strconv"
	"strings"
	"time"

	logger "github.com/ViaQ/logerr/v2/log"
	log "github.com/ViaQ/logerr/v2/log/static"
)

func init() {
	var verbosity = 0
	if level, found := os.LookupEnv("LOG_LEVEL"); found {
		if i, err := strconv.Atoi(level); err == nil {
			verbosity = i
		}
	}
	log.SetLogger(logger.NewLogger("oc-runner", logger.WithVerbosity(verbosity)))
}

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
	// never write to this channel
	return r.runCmd(make(chan time.Time, 1))
}

func (r *runner) runCmd(timeoutCh <-chan time.Time) (string, error) {
	log.V(2).Info("Running command", "cmd", CMD, "args", r.args)
	// #nosec G204
	r.Cmd = osexec.Command(CMD, r.args...)
	var outbuf bytes.Buffer
	var errbuf bytes.Buffer
	if r.tostdout {
		r.Stdout = os.Stdout
		r.Stderr = os.Stderr
	} else {
		r.Stdout = &outbuf
		r.Stderr = &errbuf
	}
	r.Env = []string{fmt.Sprintf("%s=%s", "KUBECONFIG", os.Getenv("KUBECONFIG"))}
	cmdargs := strings.Join(r.args, " ")
	err := r.Start()
	if err != nil {
		log.V(1).Error(err, "could not start oc command", "arguments", r.args, "argstr", cmdargs)
		return "", err
	}
	// Wait for the process to finish or kill it after a timeout (whichever happens first):
	done := make(chan error, 1)
	go func() {
		done <- r.Wait()
	}()
	select {
	case <-timeoutCh:
		if err = r.Process.Kill(); err != nil {
			log.V(1).Error(err, "failed to kill oc process")
		}
	case err = <-done:
		if err != nil {
			log.V(1).Error(err, "oc finished with error", "args", r.args, "argstr", cmdargs)
		}
	}
	if err != nil {
		if r.tostdout {
			return "", err
		}
		errout := strings.TrimSpace(errbuf.String())
		log.V(2).Info("command result", "arguments", r.args, "output", errout, "error", err, "argstr", cmdargs)
		return errout, err
	}
	if r.tostdout {
		return "", nil
	}
	out := strings.TrimSpace(outbuf.String())
	if len(out) > 500 {
		log.V(2).Info("output(truncated 500/length)", "arguments", r.args, "length", len(out), "result", truncateString(out, 500), "argstr", cmdargs)
	} else {
		log.V(2).Info("command output", "arguments", r.args, "output", out, "argstr", cmdargs)
	}
	return out, nil
}

func (r *runner) RunFor(d time.Duration) (string, error) {
	if r.err != nil {
		return "composed command failed", r.err
	}
	r.setArgs(r.collectArgsFunc())
	return r.runCmd(time.After(d))
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

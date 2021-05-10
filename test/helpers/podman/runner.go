package podman

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
// the methods in Command interface.
// Other commands  collect their arguments
// and use Runner to run the command with arguments.
// It provides different modes of executing the commands, Runner/RunFor/Output/OutputFor
//
// As fas as possible, it is to be kept independent of podman command syntax.
// TODO(vimalk78)
// Move KUBECONFIG out from here

// CMD is the command to be run by the runner
const CMD string = "podman"

// runner encapsulates os/exec/Cmd, collects args, and runs CMD
type runner struct {
	*osexec.Cmd

	args []string

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
	// #nosec G204
	r.Cmd = osexec.Command(CMD, r.args...)
	r.Cmd.Env = os.Environ()
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
	cmdargs := strings.Join(r.args, " ")
	err := r.Cmd.Start()
	if err != nil {
		log.Error(err, "could not start command", "cmd", r.Cmd, "arguments", cmdargs)
		return "", err
	}
	// Wait for the process to finish or kill it after a timeout (whichever happens first):
	done := make(chan error, 1)
	go func() {
		done <- r.Cmd.Wait()
	}()
	select {
	case <-timeoutCh:
		if err = r.Cmd.Process.Kill(); err != nil {
			log.Error(err, "failed to kill process: ")
		}
	case err = <-done:
		if err != nil {
			log.Error(err, "finished with error", "cmd", CMD)
		}
	}
	if err != nil {
		if r.tostdout {
			return "", err
		}
		errout := strings.TrimSpace(errbuf.String())
		log.V(1).Info("command result", "arguments", cmdargs, "output", errout, "error", err)
		return errout, err
	}
	if r.tostdout {
		return "", nil
	}
	out := strings.TrimSpace(outbuf.String())
	if len(out) > 500 {
		log.V(1).Info("output(truncated 500/length)", "arguments", cmdargs, "length", len(out), "result", truncateString(out, 500))
	} else {
		log.V(1).Info("command output", "arguments", cmdargs, "output", out)
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
	return CMD
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

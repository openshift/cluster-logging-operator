package oc

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"regexp"
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

// RunOption is a functional option for configuring Run behavior
type RunOption func(*runOptions)

// runOptions holds configuration for Run execution
type runOptions struct {
	returnRawOutput bool
}

// WithRawOutput returns the actual command output without sanitization.
// Logs will still be sanitized for security.
func WithRawOutput() RunOption {
	return func(opts *runOptions) {
		opts.returnRawOutput = true
	}
}

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

func (r *runner) Run(opts ...RunOption) (string, error) {
	if r.err != nil {
		return "composed command failed", r.err
	}
	r.setArgs(r.collectArgsFunc())

	// Apply options
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// never write to this channel
	return r.runCmd(make(chan time.Time, 1), options)
}

func (r *runner) runCmd(timeoutCh <-chan time.Time, options *runOptions) (string, error) {
	sanitizedArgs := SanitizeTokensInArgs(r.args)
	log.V(2).Info("Running command", "cmd", CMD, "args", sanitizedArgs)
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
	cmdargs := SanitizeTokenInString(strings.Join(r.args, " "))
	err := r.Start()
	if err != nil {
		log.V(1).Error(err, "could not start oc command", "arguments", sanitizedArgs, "argstr", cmdargs)
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
			log.V(1).Error(err, "oc finished with error", "args", sanitizedArgs, "argstr", cmdargs)
		}
	}
	if err != nil {
		if r.tostdout {
			return "", err
		}
		errout := strings.TrimSpace(errbuf.String())
		err = errors.Join(err, errors.New(errout))
		sanitizedErrout := SanitizeTokenInString(errout)
		log.V(2).Info("command result", "arguments", sanitizedArgs, "output", sanitizedErrout, "error", err, "argstr", cmdargs)
		return errout, err
	}
	if r.tostdout {
		return "", nil
	}
	out := strings.TrimSpace(outbuf.String())
	sanitizedOut := SanitizeTokenInString(out)
	// Log raw output at high verbosity for debugging
	log.V(9).Info("command output", "arguments", r.args, "output", out, "argstr", cmdargs)
	// Always log sanitized output for security
	if len(sanitizedOut) > 500 {
		log.V(2).Info("output(truncated 500/length)", "arguments", sanitizedArgs, "length", len(sanitizedOut), "result", truncateString(sanitizedOut, 500), "argstr", cmdargs)
	} else {
		log.V(2).Info("command output", "arguments", sanitizedArgs, "output", sanitizedOut, "argstr", cmdargs)
	}

	// Return raw or sanitized output based on options
	if options != nil && options.returnRawOutput {
		return out, nil
	}
	return sanitizedOut, nil
}

func (r *runner) RunFor(d time.Duration, opts ...RunOption) (string, error) {
	if r.err != nil {
		return "composed command failed", r.err
	}
	r.setArgs(r.collectArgsFunc())

	// Apply options
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return r.runCmd(time.After(d), options)
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

// SanitizeTokenInString replaces serviceaccount tokens with REDACTED_TOKEN
// Matches patterns like:
// - --from-literal=token=<token>
// - JWT tokens (eyJ...)
func SanitizeTokenInString(s string) string {
	jwtRegex := regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)
	return jwtRegex.ReplaceAllString(s, "[REDACTED]")
	//// Replace --from-literal=token=<value>
	//const tokenPattern = "--from-literal=token="
	//if idx := strings.Index(s, tokenPattern); idx != -1 {
	//	// Find the start of the token value
	//	start := idx + len(tokenPattern)
	//	// Find the end of the token value (next space or end of string)
	//	end := start
	//	for end < len(s) && s[end] != ' ' {
	//		end++
	//	}
	//	// Replace the token value with REDACTED_TOKEN
	//	s = s[:start] + "REDACTED_TOKEN" + s[end:]
	//}
	//
	//// Replace JWT-like tokens (base64 strings starting with eyJ)
	//// Pattern: eyJ followed by base64 characters
	//lines := strings.Split(s, "\n")
	//for j, line := range lines {
	//	words := strings.Fields(line)
	//	for i, word := range words {
	//		if strings.HasPrefix(word, "eyJ") && len(word) > 20 {
	//			words[i] = "REDACTED_TOKEN"
	//		}
	//	}
	//	lines[j] = strings.Join(words, " ")
	//}
	//return strings.Join(lines, "\n")
}

// SanitizeTokensInArgs sanitizes tokens in argument slices
func SanitizeTokensInArgs(args []string) []string {
	sanitized := make([]string, len(args))
	for i, arg := range args {
		// Check for --from-literal=token=<value>
		if strings.HasPrefix(arg, "--from-literal=token=") {
			sanitized[i] = "--from-literal=token=REDACTED_TOKEN"
		} else if strings.HasPrefix(arg, "eyJ") && len(arg) > 20 {
			// JWT token
			sanitized[i] = "REDACTED_TOKEN"
		} else {
			sanitized[i] = arg
		}
	}
	return sanitized
}

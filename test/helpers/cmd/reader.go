package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	"github.com/openshift/cluster-logging-operator/test"
)

var (
	ErrTimeout    = errors.New("timeout")
	ErrUnexpected = errors.New("unexpected data")
)

// Reader reads from a running exec.Cmd.
type Reader struct {
	cmd *exec.Cmd
	r   *bufio.Reader
}

// NewReader starts an exec.Cmd and returns a CmdReader for its stdout.
func NewReader(cmd *exec.Cmd) (*Reader, error) {
	p, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	r := &Reader{
		cmd: cmd,
		r:   bufio.NewReader(p),
	}
	return r, nil
}

// Read is the usual io.Reader, no timeout.
func (r *Reader) Read(b []byte) (int, error) { return r.r.Read(b) }

// ReadLine reads a line of text as a string. Times out after test.DefaultSuccessTimeout.
func (r *Reader) ReadLine() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), test.DefaultSuccessTimeout)
	defer cancel()
	return r.ReadLineContext(ctx)
}

// ReadLineContext returns on success or when the context is cancelled or times out.
func (r *Reader) ReadLineContext(ctx context.Context) (string, error) {
	var (
		line string
		err  error
		read = make(chan struct{})
	)
	// Read in a goroutine, abandon if timeout expires.
	go func() { line, err = r.r.ReadString('\n'); close(read) }()
	select {
	case <-read:
		if err != nil {
			if err2 := r.Close(); err2 != nil { // Extra info from process error.
				return "", fmt.Errorf("%w: %v", err, err2)
			}
			return "", err
		}
		return line, nil
	case <-ctx.Done():
		r.Close()
		return "", ctx.Err()
	}
}

// Close kills the underlying process, if running.
// This closes the stdout pipe, so Read or ReadLine will return an error.
// Returns the error returned by the sub-process.
func (r *Reader) Close() error {
	_ = r.cmd.Process.Kill()
	return r.cmd.Wait()
}

// ExpectLines calls ExpectLinesMatchContext with timeout test.DefaultSuccessTimeout.
func (r *Reader) ExpectLines(n int, good, bad string) error {
	ctx, cancel := context.WithTimeout(context.Background(), test.DefaultSuccessTimeout)
	defer cancel()
	return r.ExpectLinesContext(ctx, n, good, bad)
}

// ExpectLinesContext reads `n` lines that match regexp `good`.
// Returns error if `bad` is  not empty and a line is read that matches regexp `bad`.
// Ignores lines that do not match `good` or `bad`.
// Panics if good or bad are not valid regexps.
func (r *Reader) ExpectLinesContext(ctx context.Context, n int, good, bad string) error {
	goodx, badx := regexp.MustCompile(good), regexp.MustCompile(bad)
	for {
		line, err := r.ReadLineContext(ctx)
		switch {
		case err != nil:
			return err
		case goodx.MatchString(line):
			n--
			if n == 0 {
				return nil
			}
		case badx.String() != "" && badx.MatchString(line):
			return fmt.Errorf("bad line: %q", line)
		}
	}
}

// ExpectEmpty succeeds if nothing is read until the reader returns io.EOF or the
// context is cancelled or times out. Otherwise it returns an error.
func (r *Reader) ExpectEmpty(ctx context.Context) error {
	line, err := r.ReadLineContext(ctx)
	switch {
	case err == nil:
		return fmt.Errorf("expected empty, read line: %q", line)
	case errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		return nil
	default:
		return err
	}
}

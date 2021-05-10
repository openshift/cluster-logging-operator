package podman

import (
	"fmt"
	"time"
)

// Command is a base interface to run oc commands
type Command interface {
	fmt.Stringer

	// Runs a command, and returns command  output as result
	Run() (string, error)
	// Runs a command for the duration, and Kills the command after duration
	RunFor(time.Duration) (string, error)

	// Runs a command, sends command output to stdout
	Output() error
	// Runs a command for the duration, sends command output to stdout
	OutputFor(time.Duration) error

	// Kills a command if running
	Kill() error
}

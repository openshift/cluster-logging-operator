package oc

import (
	"fmt"
	"time"
)

// Command is a base interface to run oc commands
type Command interface {
	fmt.Stringer

	// Runs an oc command, and returns command  output as result
	Run() (string, error)
	// Runs an oc command for the duration, and Kills the command after duration
	RunFor(time.Duration) (string, error)

	// Runs an oc command, sends command output to stdout
	Output() error
	// Runs an oc command for the duration, sends command output to stdout
	OutputFor(time.Duration) error

	// Kills a command if running
	Kill() error
}

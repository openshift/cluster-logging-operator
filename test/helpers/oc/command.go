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

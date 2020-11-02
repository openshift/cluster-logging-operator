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

package matchers

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/onsi/gomega"
)

// WrapError wraps certain error types with additional information.
func WrapError(err error) error {
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) && len(exitErr.Stderr) != 0 {
		return fmt.Errorf("%w: %v", err, string(exitErr.Stderr))
	}
	return err
}

// ExpectOK is shorthand for these annoyingly long ginkgo forms:
//    Expect(err).NotTo(HaveOccured()
//    Expect(err).To(Succeed())
// It also does a WrapError to show stderr for *exec.ExitError.
func ExpectOK(err error, description ...interface{}) {
	ExpectOKWithOffset(1, err, description...)
}

func ExpectOKWithOffset(skip int, err error, description ...interface{}) {
	gomega.ExpectWithOffset(skip+1, WrapError(err)).To(gomega.Succeed(), description...)
}

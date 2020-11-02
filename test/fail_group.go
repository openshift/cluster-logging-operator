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

package test

import (
	"sync/atomic"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
)

// A FailGroup is a collection of goroutines running concurrently as part of a Ginkgo test.
// The goroutines may exit by calling ginkgo.Fail(), for example  by using Expect().
// they may also return an error.
//
// This is essentially errgroup.Group extended to recover ginkgo.Fail() correctly.
type FailGroup struct {
	g      errgroup.Group
	panics int32
}

// Go runs f concurrently and recovers from ginkgo.Fail() panics.
func (g *FailGroup) Go(f func()) { g.GoErr(func() error { f(); return nil }) }

// GoErr allows f to return an error or call ginkgo.Fail().
func (g *FailGroup) GoErr(f func() error) {
	g.g.Go(func() error {
		defer GinkgoRecover() // Recover panic and report as ginkgo test failure.
		atomic.AddInt32(&g.panics, 1)
		err := f()
		atomic.AddInt32(&g.panics, -1) // We passed f() without panic
		return err
	})
}

// Wait waits for all goroutines to exit.
//
// It will ginkgo.Fail() if any goroutine returned an error or called ginkgo.Fail().
func (g *FailGroup) Wait() {
	ExpectWithOffset(1, g.g.Wait()).To(Succeed())
	if g.panics > 0 {
		Fail("assertion failed in FailGroup", 1)
	}
}

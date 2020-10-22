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

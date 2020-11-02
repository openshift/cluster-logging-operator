package test

import (
	"fmt"

	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
)

// A FailGroup is a collection of goroutines running concurrently as part of a Ginkgo test.
// The goroutines may exit by calling ginkgo.Fail(), for example  by using Expect().
// they may also return an error.
//
// This is essentially errgroup.Group extended to recover ginkgo.Fail() correctly.
type FailGroup struct {
	g errgroup.Group
}

// Go runs f concurrently and recovers from ginkgo.Fail() panics.
func (g *FailGroup) Go(f func()) {
	g.g.Go(func() (err error) {
		defer func() {
			if v := recover(); v != nil {
				err = panicError{value: v}
			}
		}()
		f()
		return nil
	})
}

type panicError struct{ value interface{} }

func (p panicError) Error() string { return fmt.Sprintf("%v", p.value) }

// Wait waits for all goroutines to exit.
//
// It calls ginkgo.Fail() if any goroutine returned an error or called ginkgo.Fail().
func (g *FailGroup) Wait() {
	ExpectWithOffset(1, g.g.Wait()).To(Succeed())
}

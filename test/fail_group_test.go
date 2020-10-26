package test_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("FailGroup", func() {
	It("handles concurrent pass and fail", func() {
		g := test.FailGroup{}
		c := make(chan bool) // Blocking channel to ensure goroutines run concurrently
		fails := InterceptGomegaFailures(func() {
			g.Go(func() { <-c; Expect("bad").To(Equal("good")) })
			g.Go(func() { c <- true; Expect("hello").To(Equal("hello")) })
			g.Wait()
		})
		Expect(fails).To(ConsistOf("Expected\n    <string>: bad\nto equal\n    <string>: good"))
	})
})

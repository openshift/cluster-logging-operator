package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Reader", func() {
	It("reads lines", func() {
		cmd := exec.Command("echo", "a\nb\nc\n")
		r, err := NewReader(cmd)
		ExpectOK(err)
		defer r.Close()
		for _, s := range []string{"a\n", "b\n", "c\n"} {
			Expect(r.ReadLine()).To(Equal(s))
		}
	})

	It("times out", func() {
		cmd := exec.Command("sleep", "1m")
		r, err := NewReader(cmd)
		ExpectOK(err)
		defer r.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/100)
		defer cancel()
		_, err = r.ReadLineContext(ctx)
		Expect(err).To(HaveOccurred())
	})

	It("includes stderr in err.Error()", func() {
		r, err := NewReader(exec.Command("bash", "-c", "echo this is bad 1>&2; false"))
		ExpectOK(err)
		_, err = r.ReadLine()
		Expect(err).To(MatchError("EOF: exit status 1: this is bad"))
	})

	It("truncates long stderr", func() {
		long := strings.Repeat("X", 4096)
		cmd := fmt.Sprintf("echo %v 1>&2; false", long)
		r, err := NewReader(exec.Command("bash", "-c", cmd))
		ExpectOK(err)
		_, err = r.ReadLine()
		short := strings.Repeat("X", stderrLimit)
		Expect(err.Error()).To(MatchRegexp(fmt.Sprintf(" %v$", short)))
	})
})

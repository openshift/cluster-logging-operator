package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Writer", func() {
	It("writes to stdin", func() {
		cmd := exec.Command("cat")
		out := &bytes.Buffer{}
		cmd.Stdout = out
		cmd.Stderr = test.Writer()
		w, err := NewExecWriter(cmd)
		ExpectOK(err)

		_, err = fmt.Fprint(w, "hello world")
		ExpectOK(err)
		err = w.Close() // Close stdin causes cat to exit
		ExpectOK(err)
		Expect(0).To(Equal(w.ProcessState.ExitCode()))

		Expect("hello world").To(Equal(out.String()))
	})
})

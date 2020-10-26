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

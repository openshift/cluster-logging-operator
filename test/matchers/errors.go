package matchers

import (
	"github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

// ExpectOK is shorthand for these annoyingly long ginkgo forms:
//    Expect(err).NotTo(HaveOccured()
//    Expect(err).To(Succeed())
// It also does a WrapError to show stderr for *exec.ExitError.
func ExpectOK(err error, description ...interface{}) {
	ExpectOKWithOffset(1, err, description...)
}

func ExpectOKWithOffset(skip int, err error, description ...interface{}) {
	gomega.ExpectWithOffset(skip+1, utils.WrapError(err)).To(gomega.Succeed(), description...)
}

package errors_test

import (
	stderrors "errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/errors"
)

var _ = Describe("#LogIfError", func() {

	It("should return silently when no error", func() {
		Expect(func() { errors.LogIfError(nil) }).To(Not(Panic()))
	})

	Context("there is an error and messages are provided", func() {
		var (
			err = stderrors.New("some error")
		)
		It("should succeed with a single message", func() {
			Expect(func() { errors.LogIfError(err, "one") }).To(Not(Panic()))
		})
		It("should succeed with more than one message", func() {
			Expect(func() { errors.LogIfError(err, "one", "two") }).To(Not(Panic()))
		})
	})
})

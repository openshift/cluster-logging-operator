package errors

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[internal][validations][errors]", func() {

	Context("#IsValidationError", func() {

		It("should fail when not of type 'ValidationError'", func() {
			myerror := fmt.Errorf("test error")
			Expect(IsValidationError(myerror)).To(BeFalse())
		})
		It("should pass when of type 'ValidationError'", func() {
			myerror := &ValidationError{}
			Expect(IsValidationError(myerror)).To(BeTrue())
		})

	})
	Context("#MustUndeployCollector", func() {

		It("should fail when not authorized to collect", func() {
			myerror := &ValidationError{
				msg: "something" + NotAuthorizedToCollect,
			}
			Expect(MustUndeployCollector(myerror)).To(BeFalse())
		})
		It("should pass does not mention authorized to collect", func() {
			myerror := &ValidationError{
				msg: "something",
			}
			Expect(MustUndeployCollector(myerror)).To(BeTrue())
		})

	})

})

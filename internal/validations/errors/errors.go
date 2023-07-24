package errors

import (
	"fmt"
	"reflect"
)

type ValidationError struct {
	msg string
}

var validationErrorType = reflect.TypeOf(&ValidationError{})

func (e *ValidationError) Error() string {
	return e.msg
}

func NewValidationError(msg string, args ...interface{}) *ValidationError {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &ValidationError{msg: msg}
}

func IsValidationError(err error) bool {
	return reflect.TypeOf(err).AssignableTo(validationErrorType)
}

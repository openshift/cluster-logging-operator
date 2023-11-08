package helpers

import (
	"fmt"
	"reflect"
)

type OptionalPair struct {
	key   string
	Value interface{}
}

func NewOptionalPair(key string, value interface{}) OptionalPair {
	return OptionalPair{
		key,
		value,
	}
}

func (op OptionalPair) String() string {
	if op.Value == nil {
		return ""
	}
	format := "%s = %v"
	if reflect.TypeOf(op.Value).Kind() == reflect.String {
		format = "%s = %q"
	}
	return fmt.Sprintf(format, op.key, op.Value)
}

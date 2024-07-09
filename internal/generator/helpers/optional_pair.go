package helpers

import (
	"fmt"
	"reflect"

	frameworkhelper "github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

const (
	OptionFormatter = "format"
)

type OptionalPair struct {
	key     string
	Value   interface{}
	options []frameworkhelper.Option
}

func NewOptionalPair(key string, value interface{}, options ...frameworkhelper.Option) OptionalPair {
	return OptionalPair{
		key,
		value,
		options,
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
	if value, ok := frameworkhelper.HasOption(OptionFormatter, op.options); ok {
		format = value.(string)
	}
	return fmt.Sprintf(format, op.key, op.Value)
}

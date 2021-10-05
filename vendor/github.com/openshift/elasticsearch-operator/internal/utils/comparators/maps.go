package comparators

import (
	"reflect"
)

// AreStringMapsSame compares two maps which are string key/value
func AreStringMapsSame(lhs, rhs map[string]string) bool {
	return reflect.DeepEqual(lhs, rhs)
}

package v1

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"reflect"
	"strconv"
)

var (
	allowedOutputTuning = sets.NewString(
		bufferMaxEvents,
		bufferWhenFull,
		"request.concurrency",
		"request.rate_limit_duration_secs",
		"request.rate_limit_num",
		"request.retry_attempts",
		"request.retry_initial_backoff_secs",
		"request.retry_max_duration_secs",
	)
)

const (
	bufferMaxEvents = "buffer.max_events"
	bufferWhenFull  = "buffer.when_full"
)

func VariantKind(v string) reflect.Kind {
	if _, err := strconv.Atoi(string(v)); err == nil {
		return reflect.Int
	}
	return reflect.String
}
func VariantAsInt(v string) int {
	if VariantKind(v) == reflect.Int {
		if i, err := strconv.Atoi(string(v)); err == nil {
			return i
		}
	}
	return 0
}

// OutTuningSpec are additional collector implementation specific configurations
type OutputTuningSpec map[string]string

func (ot OutputTuningSpec) IsEmpty() bool {
	return len(ot) == 0
}

// Range over the tuning keys and values
func (ot OutputTuningSpec) Range(iterate func(string, string)) {
	for k, v := range ot {
		iterate(k, v)
	}
}

// Validate the spec and return the list of invalid settings
func (ot OutputTuningSpec) Validate() []error {
	errors := []error{}
	for k, v := range ot {
		if allowedOutputTuning.Has(k) {
			if !isValid(k, v) {
				errors = append(errors, fmt.Errorf("tuning value for key %q, not valid: %v", k, v))
			}
		} else {
			errors = append(errors, fmt.Errorf("tuning key not allowed: %s", k))
		}
	}
	return errors
}

func isValid(key string, value string) bool {
	switch key {
	case bufferWhenFull:
		return sets.NewString("block", "drop_newest").Has(value)
	default:
		if VariantKind(value) == reflect.Int {
			return VariantAsInt(value) > 0
		}
		return false
	}
}

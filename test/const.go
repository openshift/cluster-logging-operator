package test

import (
	"os"
	"time"
)

const (
	// DefaultSuccessTimeout for operations that are expected to succeed.
	DefaultSuccessTimeout = 10 * time.Minute

	// DefaultFailureTimeout for operations that are *expected* to time out.
	//
	// The timeout is short because timeouts are *never* a reliable way to
	// determine that something does not happen. At best such a test *might* catch
	// the problem. It is better than nothing, but not by much. Definitely not
	// worth delaying the entire test suite for.
	//
	// Ideally tests for non-happening should have a positive wait with
	// DefaultSuccessTimeout for something that we know *does* happen, and that we
	// know never happens *before* the thing we don't expect to happen.  Then look
	// for positive proof that the thing didn't happen. A negative timeout result
	// never proves that the thing won't happen in 1 more millisecond.
	//
	DefaultFailureTimeout = 1 * time.Second

	OpenshiftOperatorsRedhatNS = "openshift-operators-redhat"
)

func SuccessTimeout() time.Duration {
	d, err := time.ParseDuration(os.Getenv("SUCCESS_TIMEOUT"))
	if err != nil {
		return DefaultSuccessTimeout
	}
	return d
}

func FailureTimeout() time.Duration {
	d, err := time.ParseDuration(os.Getenv("FAILURE_TIMEOUT"))
	if err != nil {
		return DefaultFailureTimeout
	}
	return d
}

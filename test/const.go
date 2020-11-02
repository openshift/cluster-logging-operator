// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package test

import "time"

const (
	// DefaultSuccessTimeout for operations that are expected to succeed.
	DefaultSuccessTimeout = 5 * time.Minute

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

	InstanceName               = "instance"
	OpenshiftLoggingNS         = "openshift-logging"
	OpenshiftOperatorsRedhatNS = "openshift-operators-redhat"
)

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

package matchers

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/test"
	v1 "k8s.io/api/core/v1"
)

type EnvVarMatcher struct {
	expected interface{}
}

// EqualDiff is like Equal but gives cmp.Diff style output.
func IncludeEnvVar(expected interface{}) *EnvVarMatcher {
	return &EnvVarMatcher{
		expected: expected,
	}
}

func (m *EnvVarMatcher) Match(actual interface{}) (success bool, err error) {
	expVar, ok := m.expected.(v1.EnvVar)
	if !ok {
		return false, fmt.Errorf("Matcher expects v1.EnvVar")
	}
	actualVars, ok := actual.([]v1.EnvVar)
	if !ok {
		return false, fmt.Errorf("Matcher expects []v1.EnvVars")
	}
	var foundVar *v1.EnvVar
	for i := range actualVars {
		if actualVars[i].Name == expVar.Name {
			foundVar = &actualVars[i]
			break
		}
	}
	if foundVar == nil {
		return false, nil
	}
	return test.JSONString(foundVar) == test.JSONString(expVar), nil
}

func (m *EnvVarMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *EnvVarMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}

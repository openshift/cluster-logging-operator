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

package utils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestAreMapsSameWhenBothAreEmpty(t *testing.T) {
	one := map[string]string{}
	two := map[string]string{}
	if !AreMapsSame(one, two) {
		t.Error("Exp empty maps to evaluate to be equivalent")
	}
}
func TestAreMapsSameWhenOneIsEmptyAndTheOtherIsNot(t *testing.T) {
	one := map[string]string{}
	two := map[string]string{
		"foo": "bar",
	}
	if AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be different - left: %v, right: %v", one, two)
	}
}
func TestAreMapsSameWhenEquivalent(t *testing.T) {
	one := map[string]string{
		"foo": "bar",
		"xyz": "123",
	}
	two := map[string]string{
		"xyz": "123",
		"foo": "bar",
	}
	if !AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be equivalent - left: %v, right: %v", one, two)
	}
}
func TestAreMapsSameWhenDifferent(t *testing.T) {
	one := map[string]string{
		"foo": "456",
		"xyz": "123",
	}
	two := map[string]string{
		"xyz": "123",
		"foo": "bar",
	}
	if AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be different - left: %v, right: %v", one, two)
	}
}

func TestEnvVarEqualEqual(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "MERGE_JSON_LOG", Value: "false"},
	}

	if !EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned false for the equal inputs")
	}
}

func TestEnvVarEqualCheckValueFrom(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
	}

	if !EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned false for the equal inputs")
	}
}

func TestEnvVarEqualNotEqual(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "true"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}

	if EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned true for the not equal inputs")
	}
}

func TestEnvVarEqualShorter(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
	}

	if EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned true when the desired is shorter than the current")
	}
}

func TestEnvVarEqualNotEqual2(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
	}

	if EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned true when the desired is longer than the current")
	}
}

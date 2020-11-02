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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
)

var (
	limitMemory   = resource.MustParse("120Gi")
	requestMemory = resource.MustParse("100Gi")
	requestCPU    = resource.MustParse("500m")
)

func TestAreResourcesEmptyWhenUpdating(t *testing.T) {

	current := v1.ResourceRequirements{
		Limits: v1.ResourceList{v1.ResourceMemory: limitMemory},
		Requests: v1.ResourceList{
			v1.ResourceMemory: requestMemory,
			v1.ResourceCPU:    requestCPU,
		},
	}

	desired := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: requestMemory,
			v1.ResourceCPU:    requestCPU,
		},
	}

	different, result := CompareResources(current, desired)

	if !different {
		t.Error("Expected resourceRequirements to evaluate as different")
	}

	if !reflect.DeepEqual(result.Limits, desired.Limits) {
		t.Error("Expected limits to both be empty")
	}
}

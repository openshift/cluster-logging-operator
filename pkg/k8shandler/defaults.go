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

package k8shandler

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultEsMemory     resource.Quantity = resource.MustParse("16Gi")
	defaultEsCpuRequest resource.Quantity = resource.MustParse("1")

	defaultEsProxyMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultEsProxyCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultKibanaMemory     resource.Quantity = resource.MustParse("736Mi")
	defaultKibanaCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultKibanaProxyMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultKibanaProxyCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultCuratorMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultCuratorCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultFluentdMemory     resource.Quantity = resource.MustParse("736Mi")
	defaultFluentdCpuRequest resource.Quantity = resource.MustParse("100m")
)

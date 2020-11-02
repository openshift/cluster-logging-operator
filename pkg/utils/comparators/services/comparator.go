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

package services

import (
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

//AreSame compares for equality and return true equal otherwise false
func AreSame(current *v1.Service, desired *v1.Service) bool {
	log.V(3).Info("Comparing Services current to desired", "current", current, "desired", desired)

	if !utils.AreMapsSame(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
		log.V(3).Info("Service label change", "current name", current.Name)
		return false
	}
	if !utils.AreMapsSame(current.Spec.Selector, desired.Spec.Selector) {
		log.V(3).Info("Service Selector change", "current name", current.Name)
		return false
	}
	if len(current.Spec.Ports) != len(desired.Spec.Ports) {
		return false
	}
	for i, port := range current.Spec.Ports {
		dPort := desired.Spec.Ports[i]
		if !reflect.DeepEqual(port, dPort) {
			return false
		}
	}

	return true
}

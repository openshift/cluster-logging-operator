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

package daemonsets

import (
	"reflect"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	apps "k8s.io/api/apps/v1"
)

//AreSame compares daemonset for equality and return true equal otherwise false
func AreSame(current *apps.DaemonSet, desired *apps.DaemonSet) bool {
	if !utils.AreMapsSame(current.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		log.V(3).Info("DaemonSet nodeSelector change", "DaemonSetName", current.Name)
		return false
	}

	if !utils.AreTolerationsSame(current.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		log.V(3).Info("DaemonSet tolerations change", "DaemonSetName", current.Name)
		return false
	}

	if !utils.PodVolumeEquivalent(current.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		log.V(3).Info("DaemonSet volumes change", "DaemonSetName", current.Name)
		return false
	}

	if len(current.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		log.V(3).Info("DaemonSet number of containers changed", "DaemonSetName", current.Name)
		return false
	}

	if isDaemonsetImageDifference(current, desired) {
		log.V(3).Info("DaemonSet image change", "DaemonSetName", current.Name)
		return false
	}

	if utils.AreResourcesDifferent(current, desired) {
		log.V(3).Info("DaemonSet resource(s) change", "DaemonSetName", current.Name)
		return false
	}

	if !utils.EnvValueEqual(current.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env) {
		log.V(3).Info("Collector container EnvVar change found, updating ", "DaemonSetName", current.Name)
		log.V(3).Info("Collector envvars -", "current", current.Spec.Template.Spec.Containers[0].Env, "desired", desired.Spec.Template.Spec.Containers[0].Env)
		current.Spec.Template.Spec.Containers[0].Env = desired.Spec.Template.Spec.Containers[0].Env
		return false
	}

	if !reflect.DeepEqual(current.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		log.V(3).Info("Daemonset %q container volumemounts change", "DaemonSetName", current.Name)
		return false
	}

	if len(current.Spec.Template.Spec.InitContainers) != len(desired.Spec.Template.Spec.InitContainers) {
		log.V(3).Info("DaemonSet number of init containers changed", "DaemonSetName", current.Name)
		return false
	}
	for i, container := range current.Spec.Template.Spec.InitContainers {
		if container.Image != desired.Spec.Template.Spec.InitContainers[i].Image {
			log.V(3).Info("Daemonset init container image is different from desired", "DaemonSetName", current.Name, "CurrentContainerName", container.Name, "DesiredContainerName", desired.Spec.Template.Spec.InitContainers[i].Name)
			return false
		}
	}

	return true
}

func isDaemonsetImageDifference(current *apps.DaemonSet, desired *apps.DaemonSet) bool {

	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}

	return false
}

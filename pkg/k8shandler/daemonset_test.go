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
	"reflect"
	"testing"

	core "k8s.io/api/core/v1"
)

func TestNewDaemonsetDoesNotDefineMinReadySeconds(t *testing.T) {

	podspec := core.PodSpec{}
	daemonSet := NewDaemonSet("thenname", "thenamespace", "thecomponent", "thecomponent", podspec)

	if daemonSet.Spec.MinReadySeconds != 0 {
		t.Errorf("Exp. the MinReadySeconds to be the default but was %d", daemonSet.Spec.MinReadySeconds)
	}
}
func TestNewDaemonsetSetsAllLabelsToBeTheSame(t *testing.T) {

	podspec := core.PodSpec{}
	daemonSet := NewDaemonSet("thenname", "thenamespace", "thecomponent", "thecomponent", podspec)

	expLabels := daemonSet.ObjectMeta.Labels
	if !reflect.DeepEqual(expLabels, daemonSet.Spec.Selector.MatchLabels) {
		t.Errorf("Exp. the ObjectMeta.Labels %q to be the same as spec.selector.matchlabels: %q", expLabels, daemonSet.Spec.Selector.MatchLabels)
	}
	if !reflect.DeepEqual(expLabels, daemonSet.Spec.Template.ObjectMeta.Labels) {
		t.Errorf("Exp. the ObjectMeta.Labels %q to be the same as spec.template.objectmeta.labels: %q", expLabels, daemonSet.Spec.Selector.MatchLabels)
	}
}
func TestNewDaemonsetIncludesCriticalPodAnnotation(t *testing.T) {

	podspec := core.PodSpec{}
	daemonSet := NewDaemonSet("thenname", "thenamespace", "thecomponent", "thecomponent", podspec)

	if _, ok := daemonSet.Spec.Template.ObjectMeta.Annotations["scheduler.alpha.kubernetes.io/critical-pod"]; !ok {
		t.Error("Exp. the daemonset to define the critical pod annotation but it did not")
	}
}

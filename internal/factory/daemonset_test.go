package factory

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
	for k, v := range daemonSet.Spec.Selector.MatchLabels {
		if _, ok := expLabels[k]; !ok {
			t.Errorf("spec.selector.MatchLabel key: %q does not exist in ObjectMeta.Labels", k)
		}
		if !reflect.DeepEqual(expLabels[k], v) {
			t.Errorf("spec.selector.MatchLabel[%q] value %v is not same as ObjectMeta.Labels[%q]: %v", k, expLabels[k], k, v)
		}
	}
	for k, v := range daemonSet.Spec.Template.ObjectMeta.Labels {
		if _, ok := expLabels[k]; !ok {
			t.Errorf("spec.template.ObjectMeta.Labels key: %q does not exist in ObjectMeta.Labels", k)
		}
		if !reflect.DeepEqual(expLabels[k], v) {
			t.Errorf("spec.template.ObjectMeta.Labels[%q] value %v is not same as ObjectMeta.Labels[%q]: %v", k, expLabels[k], k, v)
		}
	}
}
func TestNewDaemonsetIncludesCriticalPodAnnotation(t *testing.T) {

	podspec := core.PodSpec{}
	daemonSet := NewDaemonSet("thenname", "thenamespace", "thecomponent", "thecomponent", podspec)

	if _, ok := daemonSet.Spec.Template.ObjectMeta.Annotations["scheduler.alpha.kubernetes.io/critical-pod"]; !ok {
		t.Error("Exp. the daemonset to define the critical pod annotation but it did not")
	}
}

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

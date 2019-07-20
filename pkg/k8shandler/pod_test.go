package k8shandler

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

// Call this method if you want to test that podSpec's node selectors contain only the linux one.
// Note: This method gets called from other tests. That is why we made it public to other packages.
func CheckIfThereIsOnlyTheLinuxSelector(podSpec core.PodSpec, t *testing.T) {
	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contains the linux allocation selector but was %T", podSpec.NodeSelector)
	}
	if len(podSpec.NodeSelector) != 1 {
		t.Errorf("Exp. single nodeSelector but %d were found", len(podSpec.NodeSelector))
	}
	if podSpec.NodeSelector[utils.OsNodeLabel] != utils.LinuxValue {
		t.Errorf("Exp. the nodeSelector to contains %s: %s pair", utils.OsNodeLabel, utils.LinuxValue)
	}
}

func TestNodeAllocationLabelsForPod(t *testing.T) {

	// Create pod with nil selectors, we expect a new selectors map will be created
	// and it will contain only linux allocation selector.
	podSpec := NewPodSpec(
		"Foo",
		[]v1.Container{},
		[]v1.Volume{},
		nil,
		nil,
	)

	CheckIfThereIsOnlyTheLinuxSelector(podSpec, t)

	// Create pod with some "foo" selector, we expect a new linux box selector will be added
	// while existing selectors will be left intact.
	podSpec = NewPodSpec(
		"Foo",
		[]v1.Container{},
		[]v1.Volume{},
		map[string]string{"foo": "bar"},
		nil,
	)

	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contains the linux allocation selector but was %T", podSpec.NodeSelector)
	}
	if len(podSpec.NodeSelector) != 2 {
		t.Errorf("Exp. single nodeSelector but %d were found", len(podSpec.NodeSelector))
	}
	if podSpec.NodeSelector["foo"] != "bar" {
		t.Errorf("Exp. the nodeSelector to contains %s: %s pair", "foo", "bar")
	}
	if podSpec.NodeSelector[utils.OsNodeLabel] != utils.LinuxValue {
		t.Errorf("Exp. the nodeSelector to contains %s: %s pair", utils.OsNodeLabel, utils.LinuxValue)
	}

	// Create pod with "linux" selector, we expect it stays unchanged.
	podSpec = NewPodSpec(
		"Foo",
		[]v1.Container{},
		[]v1.Volume{},
		map[string]string{utils.OsNodeLabel: utils.LinuxValue},
		nil,
	)

	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contains the linux allocation selector but was %T", podSpec.NodeSelector)
	}
	if len(podSpec.NodeSelector) != 1 {
		t.Errorf("Exp. single nodeSelector but %d were found", len(podSpec.NodeSelector))
	}
	if podSpec.NodeSelector[utils.OsNodeLabel] != utils.LinuxValue {
		t.Errorf("Exp. the nodeSelector to contains %s: %s pair", utils.OsNodeLabel, utils.LinuxValue)
	}

	// Create pod with some "non-linux" selector, we expect it is overridden.
	podSpec = NewPodSpec(
		"Foo",
		[]v1.Container{},
		[]v1.Volume{},
		map[string]string{utils.OsNodeLabel: "Donald Duck"},
		nil,
	)

	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contains the linux allocation selector but was %T", podSpec.NodeSelector)
	}
	if len(podSpec.NodeSelector) != 1 {
		t.Errorf("Exp. single nodeSelector but %d were found", len(podSpec.NodeSelector))
	}
	if custom := podSpec.NodeSelector[utils.OsNodeLabel]; custom != utils.LinuxValue {
		t.Errorf("Exp. the nodeSelector was overridden from %s: %s pair to %s", utils.OsNodeLabel, utils.LinuxValue, custom)
	}
}

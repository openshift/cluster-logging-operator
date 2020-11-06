package k8shandler

import (
	"reflect"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/openshift/elasticsearch-operator/pkg/utils/comparators"
	v1 "k8s.io/api/core/v1"
)

// ArePodTemplateSpecDifferent compares two v1.PodTemplateSpecs
// and returns True or False
func ArePodTemplateSpecDifferent(lhs, rhs v1.PodTemplateSpec) bool {

	return ArePodSpecDifferent(lhs.Spec, rhs.Spec, true)
}

// Abstracted logic into comparing pod specs so that we can check if our change has been rolled out
// yet or not
func ArePodSpecDifferent(lhs, rhs v1.PodSpec, strictTolerations bool) bool {
	changed := false

	if len(lhs.Containers) != len(rhs.Containers) {
		changed = true
	}

	// check nodeselectors
	if !areSelectorsSame(lhs.NodeSelector, rhs.NodeSelector) {
		//logrus.Debugf("Resource '%s' has different nodeSelector than desired", node.self.Name)
		changed = true
	}

	// check if volumes are the same
	if !containsSameVolumes(lhs.Volumes, rhs.Volumes) {
		changed = true
	}

	// strictTolerations are for when we compare from the deployments or statefulsets
	// if we are seeing if rolled out pods contain changes we don't want strictTolerations
	//   since k8s may add additional tolerations to pods
	if strictTolerations {
		// check tolerations
		if !areTolerationsSame(lhs.Tolerations, rhs.Tolerations) {
			//logrus.Debugf("Resource '%s' has different tolerations than desired", node.self.Name)
			changed = true
		}
	} else {
		// check tolerations
		if !containsSameTolerations(lhs.Tolerations, rhs.Tolerations) {
			//logrus.Debugf("Resource '%s' has different tolerations than desired", node.self.Name)
			changed = true
		}
	}

	// check container fields
	for _, lContainer := range lhs.Containers {
		found := false

		for _, rContainer := range rhs.Containers {
			// Only compare containers with the same name
			if lContainer.Name != rContainer.Name {
				continue
			}

			found = true

			// can't use reflect.DeepEqual here, due to k8s adding token mounts
			// check that rContainer is all found within lContainer and that they match by name
			if !containsSameVolumeMounts(lContainer.VolumeMounts, rContainer.VolumeMounts) {
				changed = true
			}

			if lContainer.Image != rContainer.Image {
				changed = true
			}

			if !comparators.EnvValueEqual(lContainer.Env, rContainer.Env) {
				changed = true
			}

			if !reflect.DeepEqual(lContainer.Args, rContainer.Args) {
				changed = true
			}

			if !reflect.DeepEqual(lContainer.Ports, rContainer.Ports) {
				changed = true
			}

			if different, _ := utils.CompareResources(lContainer.Resources, rContainer.Resources); different {
				changed = true
			}
		}

		if !found {
			changed = true
		}
	}
	return changed
}

// check that all of rhs (desired) are contained within lhs (current)
func containsSameVolumeMounts(lhs, rhs []v1.VolumeMount) bool {

	for _, rVolumeMount := range rhs {
		found := false

		for _, lVolumeMount := range lhs {
			if lVolumeMount.Name == rVolumeMount.Name {
				found = true

				if !reflect.DeepEqual(lVolumeMount, rVolumeMount) {
					return false
				}
			}
		}

		if !found {
			return false
		}
	}

	return true
}

// if we use reflect.DeepEqual we will keep recognizing a difference due to defaultModes
// we want to check that rhs is contained within lhs
func containsSameVolumes(lhs, rhs []v1.Volume) bool {

	for _, rVolume := range rhs {
		found := false

		for _, lVolume := range lhs {
			if lVolume.Name == rVolume.Name {

				found = true

				if lVolume.ConfigMap != nil || rVolume.ConfigMap != nil {
					if rVolume.ConfigMap == nil {
						return false
					}

					if lVolume.ConfigMap == nil {
						return false
					}

					if lVolume.ConfigMap.Name != rVolume.ConfigMap.Name {
						return false
					}
				}

				if lVolume.Secret != nil || rVolume.Secret != nil {
					if rVolume.Secret == nil {
						return false
					}

					if lVolume.Secret == nil {
						return false
					}

					if lVolume.Secret.SecretName != rVolume.Secret.SecretName {
						return false
					}
				}

				if lVolume.PersistentVolumeClaim != nil || rVolume.PersistentVolumeClaim != nil {
					if rVolume.PersistentVolumeClaim == nil {
						return false
					}

					if lVolume.PersistentVolumeClaim == nil {
						return false
					}

					if lVolume.PersistentVolumeClaim.ClaimName != rVolume.PersistentVolumeClaim.ClaimName {
						return false
					}
				}

				if lVolume.EmptyDir != nil || rVolume.EmptyDir != nil {
					if rVolume.EmptyDir == nil {
						return false
					}

					if lVolume.EmptyDir == nil {
						return false
					}

					if lVolume.EmptyDir.SizeLimit != nil || rVolume.EmptyDir.SizeLimit != nil {
						if rVolume.EmptyDir.SizeLimit == nil {
							return false
						}

						if lVolume.EmptyDir.SizeLimit == nil {
							return false
						}

						if *lVolume.EmptyDir.SizeLimit != *rVolume.EmptyDir.SizeLimit {
							return false
						}
					}
				}
			}
		}

		if !found {
			return false
		}
	}

	return true
}

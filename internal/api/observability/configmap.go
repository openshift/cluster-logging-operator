package observability

import corev1 "k8s.io/api/core/v1"

type ConfigMaps map[string]*corev1.ConfigMap

func (c ConfigMaps) Names() (names []string) {
	for name := range c {
		names = append(names, name)
	}
	return names
}

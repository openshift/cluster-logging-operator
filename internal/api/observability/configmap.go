package observability

import (
	"fmt"
	"hash/fnv"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

type ConfigMaps map[string]*corev1.ConfigMap

// Hash64a returns an FNV-1a representation of the configmaps
func (c ConfigMaps) Hash64a() string {
	names := c.Names()
	buffer := fnv.New64a()
	for _, name := range names {
		cm := c[name]
		buffer.Write([]byte(name))

		var keys []string
		for key := range cm.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := cm.Data[k]
			buffer.Write([]byte(k))
			buffer.Write([]byte(v))
		}
	}
	return fmt.Sprintf("%d", buffer.Sum64())
}

func (c ConfigMaps) Names() (names []string) {
	for name := range c {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

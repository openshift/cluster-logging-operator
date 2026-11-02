package observability

import (
	"fmt"
	"hash/fnv"
	"path/filepath"
	"sort"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
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

// Path returns the path to the given secret key if it exists or empty
func (s ConfigMaps) Path(key *obs.ValueReference, formatter ...string) string {
	if key.ConfigMapName != "" && s[key.ConfigMapName] != nil {
		return ConfigPath(key.ConfigMapName, key.Key, formatter...)
	}
	return ""
}

// ConfigPath is the quoted path for any configmap visible to the collector
func ConfigPath(name string, file string, formatter ...string) string {
	formatString := "%q"
	if len(formatter) > 0 {
		formatString = formatter[0]
	}
	return fmt.Sprintf(formatString, filepath.Join(constants.ConfigMapBaseDir, name, file))
}

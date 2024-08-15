package collector

import (
	"fmt"
	"hash/fnv"
	v1 "k8s.io/api/core/v1"
	"sort"
)

type Secrets map[string]*v1.Secret

func (s Secrets) Names() (names []string) {

	for name := range s {
		names = append(names, name)
	}
	return names
}

// SecretsHash64a returns an FNV-1a representation of a secret map
func SecretsHash64a(s Secrets) string {
	names := s.Names()
	sort.Strings(names)
	buffer := fnv.New64a()
	for _, name := range names {
		secret := s[name]
		buffer.Write([]byte(name))

		var keys []string
		for key := range secret.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := secret.Data[k]
			buffer.Write([]byte(k))
			buffer.Write(v)
		}
	}
	return fmt.Sprintf("%d", buffer.Sum64())
}

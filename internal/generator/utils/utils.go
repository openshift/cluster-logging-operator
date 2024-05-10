package utils

import (
	"fmt"
	"sort"
	"strings"
)

// TODO: DELETE ME once fluentd removal merges
func ToHeaderStr(h map[string]string, formatStr string) string {
	if len(h) == 0 {
		return ""
	}
	sortedKeys := make([]string, len(h))
	i := 0
	for k := range h {
		sortedKeys[i] = k
		i += 1
	}
	sort.Strings(sortedKeys)
	hv := make([]string, len(h))
	for i, k := range sortedKeys {
		hv[i] = fmt.Sprintf(formatStr, k, h[k])
	}
	return fmt.Sprintf("{%s}", strings.Join(hv, ","))
}

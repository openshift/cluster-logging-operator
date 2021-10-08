package helpers

import (
	"fmt"
	"strings"
)

var Replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")

func StoreID(prefix, name, suffix string) string {
	return strings.ToLower(fmt.Sprintf("%v%v%v", prefix, Replacer.Replace(name), suffix))
}

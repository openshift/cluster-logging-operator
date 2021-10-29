package helpers

import (
	"fmt"
	"strings"
)

func MakeInputs(in ...string) string {
	out := make([]string, len(in))
	for i, o := range in {
		if strings.HasPrefix(o, "\"") && strings.HasSuffix(o, "\"") {
			out[i] = o
		} else {
			out[i] = fmt.Sprintf("%q", o)
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(out, ","))
}

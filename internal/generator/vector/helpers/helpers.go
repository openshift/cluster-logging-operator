package helpers

import (
	"fmt"
	"strings"
)

func PipelineName(name string) string {
	return strings.ToLower(name)
}

func ConcatArrays(input []string) string {
	out := make([]string, len(input))
	for i, a := range input {
		out[i] = fmt.Sprintf("%q", a)
	}
	return fmt.Sprintf("[%s]", strings.Join(out, ", "))
}

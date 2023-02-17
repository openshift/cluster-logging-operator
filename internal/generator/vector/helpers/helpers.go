package helpers

import (
	"fmt"
	"strings"
)

var Replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")

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

func TrimSpaces(in []string) []string {
	o := make([]string, len(in))
	for i, s := range in {
		o[i] = strings.TrimSpace(s)
	}
	return o
}

func FormatComponentID(name string) string {
	return strings.ToLower(Replacer.Replace(name))
}

func FormatConf(conf string) string {
	lines := strings.Split(conf, "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		trimmed = strings.ReplaceAll(trimmed, "\t", "  ")
		if len(strings.TrimSpace(trimmed)) == 0 {
			trimmed = ""
		}
		lines[i] = trimmed
	}
	return strings.Join(lines, "\n")
}

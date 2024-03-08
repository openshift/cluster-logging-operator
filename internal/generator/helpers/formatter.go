package helpers

import (
	"strings"
)

func FormatFluentConf(conf string) string {
	indent := 0
	lines := strings.Split(conf, "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(trimmed, "</") && strings.HasSuffix(trimmed, ">"):
			indent--
			trimmed = pad(trimmed, indent)
		case strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">"):
			trimmed = pad(trimmed, indent)
			indent++
		default:
			trimmed = pad(trimmed, indent)
		}
		if len(strings.TrimSpace(trimmed)) == 0 {
			trimmed = ""
		}
		lines[i] = trimmed
	}
	return strings.Join(lines, "\n")
}

func pad(line string, indent int) string {
	prefix := ""
	if indent >= 0 {
		prefix = strings.Repeat("  ", indent)
	}
	return prefix + line
}

func FormatVectorToml(conf string) string {
	result := []string{}
	prev := ""
	for _, line := range strings.Split(conf, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			switch {
			case prev == "" && strings.HasPrefix(trimmed, "#"):
				result = append(result, "")
			case !strings.HasPrefix(prev, "#") && strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"):
				result = append(result, "")
			}
			result = append(result, line)
		}
		prev = line
	}
	return strings.Join(result, "\n")
}

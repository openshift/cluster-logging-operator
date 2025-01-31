package helpers

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

var (
	Replacer         = strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	listenAllAddress string
	listenAllOnce    sync.Once
	PathRegex        = regexp.MustCompile(`\{([^{}]+)\}`)
	splitRegex       = regexp.MustCompile(`to_string!\(([^)]+)\)`)
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

func ListenOnAllLocalInterfacesAddress() string {
	f := func() {
		listenAllAddress = func() string {
			if fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, unix.IPPROTO_IP); err != nil {
				return `0.0.0.0`
			} else {
				unix.Close(fd)
				return `[::]`
			}
		}()
	}
	listenAllOnce.Do(f)
	return listenAllAddress
}

func EscapeDollarSigns(s string) string {
	return strings.ReplaceAll(s, "$", "$$")
}

// TransformUserTemplateToVRL converts the user entered template to VRL compatible syntax
// Example: foo-{.log_type||"none"} -> "foo-" + to_string!(.log_type||"none")
func TransformUserTemplateToVRL(userTemplate string) string {
	// Finds and replaces expressions defined in `{}` with to_string!()
	replacedUserTemplate := ReplaceBracketWithToString(userTemplate, "to_string!(%s)")

	// Finding all matches of to_string!() returning their start + end indices
	matchedIndices := splitRegex.FindAllStringSubmatchIndex(replacedUserTemplate, -1)
	if len(matchedIndices) == 0 {
		return fmt.Sprintf("%q", userTemplate)
	}

	var result []string
	lastIndex := 0
	// Make the final resulting array with the appropriate pieces so that it can be concatenated together with `+`
	for _, match := range matchedIndices {
		// Append the part before the match. Check if empty string so we don't concat it
		if beforePart := replacedUserTemplate[lastIndex:match[0]]; beforePart != "" {
			result = append(result, fmt.Sprintf("%q", beforePart))
		}

		// Append the to_string!() group and replace any labels with values from internal context
		result = append(result, strings.NewReplacer(
			".kubernetes.labels", "._internal.kubernetes.labels",
			".kubernetes.namespace_labels", "._internal.kubernetes.namespace_labels",
			".openshift.labels", "._internal.openshift.labels",
		).Replace(replacedUserTemplate[match[0]:match[1]]))
		lastIndex = match[1]
	}
	// Append the remaining part of the string after the last match making sure it isn't the empty string
	if endPart := replacedUserTemplate[lastIndex:]; endPart != "" {
		result = append(result, fmt.Sprintf("%q", endPart))
	}

	// Join array with `+`
	return strings.Join(result, " + ")
}

func ReplaceBracketWithToString(userTemplate, replaceWith string) string {
	return PathRegex.ReplaceAllStringFunc(userTemplate, func(match string) string {
		matches := PathRegex.FindStringSubmatch(match)
		replaced := fmt.Sprintf(replaceWith, matches[1])
		return replaced
	})
}

package template

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"text/template"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

var (
	//go:embed template.vrl.tmpl
	templateVRLTmplStr  string
	PathRegex           = regexp.MustCompile(`\{([^{}]+)\}`)
	splitRegex          = regexp.MustCompile(`to_string!\(([^)]+)\)`)
	UserTemplateVRLTmpl = template.Must(template.New("template VRL").Parse(templateVRLTmplStr))
)

type Template struct {
	Field     string
	VRLString string
}

func TemplateRemap(componentID string, inputs []string, userTemplate, field, description string) framework.Element {
	// Generate template
	w := &strings.Builder{}

	_ = UserTemplateVRLTmpl.Execute(w,
		Template{
			Field:     field,
			VRLString: TransformUserTemplateToVRL(userTemplate),
		},
	)

	return elements.Remap{
		Desc:        description,
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         w.String(),
	}
}

// TransformUserTemplateToVRL converts the user entered template to VRL compatible syntax
// Example: foo-{.log_type||"none"} -> "foo-" + to_string!(._internal.log_type||"none")
// Also supports timestamp patterns: {@timestamp|date} -> format_timestamp!(.timestamp || now(), format: "%Y-%m-%d")
func TransformUserTemplateToVRL(userTemplate string) string {
	// Check if this template contains timestamp patterns
	hasTimestampPatterns := strings.Contains(userTemplate, "@timestamp|")

	if hasTimestampPatterns {
		return transformWithTimestampSupport(userTemplate)
	}

	// Use the original logic for backwards compatibility
	return transformOriginal(userTemplate)
}

// TODO: refactor    (without a map)
// transformWithTimestampSupport handles templates with timestamp patterns
func transformWithTimestampSupport(userTemplate string) string {
	// Process the template by replacing patterns in order
	result := userTemplate

	// First pass: replace timestamp patterns
	timestampPatterns := map[string]string{
		`\{@timestamp\|strftime:"([^"]+)"\}`: `format_timestamp!(.timestamp || now(), format: "$1")`,
		`\{@timestamp\|year\}`:               `format_timestamp!(.timestamp || now(), format: "%Y")`,
		`\{@timestamp\|month\}`:              `format_timestamp!(.timestamp || now(), format: "%m")`,
		`\{@timestamp\|day\}`:                `format_timestamp!(.timestamp || now(), format: "%d")`,
		`\{@timestamp\|hour\}`:               `format_timestamp!(.timestamp || now(), format: "%H")`,
		`\{@timestamp\|minute\}`:             `format_timestamp!(.timestamp || now(), format: "%M")`,
		`\{@timestamp\|date\}`:               `format_timestamp!(.timestamp || now(), format: "%Y-%m-%d")`,
		`\{@timestamp\|datetime\}`:           `format_timestamp!(.timestamp || now(), format: "%Y-%m-%d_%H-%M-%S")`,
	}

	for pattern, replacement := range timestampPatterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, replacement)
	}

	// Second pass: replace field patterns (but not inside already processed functions)
	result = PathRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Skip if this is part of an already processed function
		if strings.Contains(match, "format_timestamp!") || strings.Contains(match, "to_string!") {
			return match
		}

		matches := PathRegex.FindStringSubmatch(match)
		fieldPath := matches[1]
		return fmt.Sprintf("to_string!(._internal%s)", fieldPath)
	})

	// Third pass: split into components for VRL concatenation
	// Parse function calls manually to handle nested parentheses properly
	var parts []string
	i := 0

	for i < len(result) {
		// Look for the start of a function call
		if start := strings.Index(result[i:], "to_string!("); start != -1 {
			start += i
			// Add text before function if any
			if beforeText := result[i:start]; beforeText != "" {
				parts = append(parts, fmt.Sprintf("%q", beforeText))
			}

			// Find the matching closing parenthesis
			end := findMatchingParen(result, start+10) // 10 = len("to_string!(") - 1
			parts = append(parts, result[start:end+1])
			i = end + 1
			continue
		}

		if start := strings.Index(result[i:], "format_timestamp!("); start != -1 {
			start += i
			// Add text before function if any
			if beforeText := result[i:start]; beforeText != "" {
				parts = append(parts, fmt.Sprintf("%q", beforeText))
			}

			// Find the matching closing parenthesis
			end := findMatchingParen(result, start+17) // 17 = len("format_timestamp!(") - 1
			parts = append(parts, result[start:end+1])
			i = end + 1
			continue
		}

		// No more functions, add remaining text
		if remaining := result[i:]; remaining != "" {
			parts = append(parts, fmt.Sprintf("%q", remaining))
		}
		break
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%q", userTemplate)
	}

	return strings.Join(parts, " + ")
}

// transformOriginal uses the original logic for backwards compatibility
func transformOriginal(userTemplate string) string {
	// Finds and replaces expressions defined in `{}` with to_string!()
	replacedUserTemplate := ReplaceBracketWithToString(userTemplate, "to_string!(._internal%s)")

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

		result = append(result, replacedUserTemplate[match[0]:match[1]])
		lastIndex = match[1]
	}
	// Append the remaining part of the string after the last match making sure it isn't the empty string
	if endPart := replacedUserTemplate[lastIndex:]; endPart != "" {
		result = append(result, fmt.Sprintf("%q", endPart))
	}

	// Join array with `+`
	return strings.Join(result, " + ")
}

// findMatchingParen finds the index of the matching closing parenthesis
func findMatchingParen(s string, start int) int {
	if start >= len(s) || s[start] != '(' {
		return start
	}

	count := 1
	inQuotes := false
	escapeNext := false

	for i := start + 1; i < len(s); i++ {
		if escapeNext {
			escapeNext = false
			continue
		}

		if s[i] == '\\' {
			escapeNext = true
			continue
		}

		if s[i] == '"' {
			inQuotes = !inQuotes
			continue
		}

		if !inQuotes {
			if s[i] == '(' {
				count++
			} else if s[i] == ')' {
				count--
				if count == 0 {
					return i
				}
			}
		}
	}

	return len(s) - 1 // fallback
}

func ReplaceBracketWithToString(userTemplate, replaceWith string) string {
	return PathRegex.ReplaceAllStringFunc(userTemplate, func(match string) string {
		matches := PathRegex.FindStringSubmatch(match)
		replaced := fmt.Sprintf(replaceWith, matches[1])
		return replaced
	})
}

package prune

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

type Prune struct {
	In    string
	NotIn string
}

var (
	// Match quoted strings like "foo" or "foo/bar-baz"
	quoteRegex       = regexp.MustCompile(`^".+"$`)
	PruneVRLTemplate = template.Must(template.New("prune VRL").Parse(pruneVRLTemplateStr))

	//go:embed prune.vrl.tmpl
	pruneVRLTemplateStr string

	dedottedFields = []string{".kubernetes.labels.", ".kubernetes.namespace_labels."}
)

type PruneFilter obs.PruneFilterSpec

func NewFilter(pruneFilterSpec *obs.PruneFilterSpec) PruneFilter {
	return PruneFilter(*pruneFilterSpec)
}

func (f PruneFilter) VRL() (string, error) {
	Prune := Prune{}
	if f.NotIn != nil {
		Prune.NotIn = generateQuotedPathSegmentArrayStr(f.NotIn)
	}
	if f.In != nil {
		Prune.In = generateQuotedPathSegmentArrayStr(f.In)
	}

	// Execute Go template to generate VRL
	w := &strings.Builder{}
	err := PruneVRLTemplate.Execute(w, Prune)
	return w.String(), err
}

// generateQuotedPathSegmentArrayStr generates the final string of the array of array of path segments
// to feed into VRL
func generateQuotedPathSegmentArrayStr(fieldPathArray []obs.FieldPath) string {
	quotedPathArray := []string{}

	rootPaths := [][]string{
		{`"_internal"`},
		{`"_internal"`, `"structured"`},
	}

	formatSegments := func(path obs.FieldPath, root []string) string {
		splitPathSegments := splitPath(string(path))
		pathArray := append([]string{}, root...)
		pathArray = append(pathArray, quotePathSegments(splitPathSegments)...)

		return fmt.Sprintf("[%s]", strings.Join(pathArray, ","))
	}

	for _, fieldPath := range fieldPathArray {
		for _, root := range rootPaths {
			quotedPathArray = append(quotedPathArray, formatSegments(fieldPath, root))
		}

		for _, d := range dedottedFields {
			label, found := strings.CutPrefix(string(fieldPath), d)
			if found && strings.ContainsAny(label, "/.") {
				label = strings.ReplaceAll(label, ".", "_")
				label = strings.ReplaceAll(label, "/", "_")

				transformedPath := obs.FieldPath(d + label)

				for _, root := range rootPaths {
					quotedPathArray = append(quotedPathArray, formatSegments(transformedPath, root))
				}
			}
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(quotedPathArray, ","))
}

// splitPath splits a fieldPath by `.` and reassembles the quoted path segments that also contain `.`
// Example: `.foo."@some"."d.f.g.o111-22/333".foo_bar`
// Resultant Array: ["foo","@some",`"d.f.g.o111-22/333"`,"foo_bar"]
func splitPath(path string) []string {
	result := []string{}

	splitPath := strings.Split(path, ".")

	var currSegment string
	for _, part := range splitPath {
		if part == "" {
			continue
		} else if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			result = append(result, part)
		} else if strings.HasPrefix(part, `"`) {
			currSegment = part
		} else if strings.HasSuffix(part, `"`) {
			currSegment += "." + part
			result = append(result, currSegment)
			currSegment = ""
		} else if currSegment != "" {
			currSegment += "." + part
		} else {
			result = append(result, part)
		}
	}
	return result
}

// quotePathSegments quotes all path segments as needed for VRL
func quotePathSegments(pathArray []string) []string {
	for i, field := range pathArray {
		// Don't surround in quotes if already quoted
		if quoteRegex.MatchString(field) {
			continue
		}
		// Put quotes around path segments
		pathArray[i] = fmt.Sprintf("%q", field)
	}
	return pathArray
}

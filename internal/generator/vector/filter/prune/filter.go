package prune

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
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
)

func MakePruneFilter(pruneFilterSpec *loggingv1.PruneFilterSpec) (vrl string, err error) {
	Prune := Prune{}
	if pruneFilterSpec.NotIn != nil {
		Prune.NotIn = generateQuotedPathSegmentArrayStr(pruneFilterSpec.NotIn)
	}
	if pruneFilterSpec.In != nil {
		Prune.In = generateQuotedPathSegmentArrayStr(pruneFilterSpec.In)
	}

	// Execute Go template to generate VRL
	w := &strings.Builder{}
	err = PruneVRLTemplate.Execute(w, Prune)
	return w.String(), err
}

// generateQuotedPathSegmentArrayStr generates the final string of the array of array of path segments
// to feed into VRL
func generateQuotedPathSegmentArrayStr(fieldPathArray []string) string {
	quotedPathArray := []string{}
	for _, fieldPath := range fieldPathArray {
		pathArray := quotePathSegments(strings.Split(fieldPath, ".")[1:])
		quotedPathArray = append(quotedPathArray, fmt.Sprintf("[%s]", strings.Join(pathArray, ",")))
	}
	return fmt.Sprintf("[%s]", strings.Join(quotedPathArray, ","))
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

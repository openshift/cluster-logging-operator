package apiaudit

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// matchAny returns an anchored regexp that matches any of the wildcard strings.
func matchAny(wildcards []string) string {
	w := &strings.Builder{}
	anyWildcardRegexp(w, wildcards, "")
	return w.String()
}

// matchAnyPath returns an anchored regexp that matches an optional trailing "?.*"
// to match against the path part of a URI path with a query.
func matchAnyPath(wildcards []string) string {
	w := &strings.Builder{}
	anyWildcardRegexp(w, wildcards, `(\?.*)?`)
	return w.String()
}

// anyWildcardRegexp returns an anchored regexp matching any of the given wildcards.
// The suffix is appended to the regex for each wildcard.
func anyWildcardRegexp(w io.Writer, wildcards []string, suffix string) {
	fmt.Fprint(w, `r'^(`)
	for i, wildcard := range wildcards {
		if i > 0 {
			fmt.Fprint(w, "|")
		}
		wildcardRegexp(w, wildcard)
	}
	fmt.Fprint(w, `)`)
	fmt.Fprint(w, suffix)
	fmt.Fprint(w, `$'`)
}

// wildcardRegexp writes an un-anchroed regexp equivalent to wildcard with leading or trailing "*"
func wildcardRegexp(w io.Writer, wildcard string) {
	if strings.HasPrefix(wildcard, "*") {
		fmt.Fprint(w, ".*")
		wildcard = wildcard[1:]
	}
	if strings.HasSuffix(wildcard, "*") {
		fmt.Fprint(w, regexp.QuoteMeta(wildcard[:len(wildcard)-1]))
		fmt.Fprint(w, ".*")
	} else {
		fmt.Fprint(w, regexp.QuoteMeta(wildcard))
	}
}

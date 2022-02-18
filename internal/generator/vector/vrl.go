package vector

import (
	"fmt"
	"strings"
)

var (
	SkipEmpty = func(in []string) []string {
		vals := []string{}
		for _, v := range in {
			if strings.TrimSpace(v) != "" {
				vals = append(vals, v)
			}
		}
		return vals
	}
	Paren = func(in string) string {
		return fmt.Sprintf("(%s)", in)
	}
	ParenAll = func(in []string) []string {
		if len(in) == 1 {
			return in
		}
		vals := []string{}
		for _, v := range in {
			vals = append(vals, Paren(v))
		}
		return vals
	}
	StartWith = func(x, y string) string {
		return fmt.Sprintf("starts_with!(%s,%q)", x, y)
	}
	Eq = func(x, y string) string {
		return fmt.Sprintf("%s == %q", x, y)
	}
	Quote = func(expr string) string {
		return fmt.Sprintf("'%s'", expr)
	}
	OR = func(nsExpr ...string) string {
		return strings.Join(ParenAll(SkipEmpty(nsExpr)), " || ")
	}
	AND = func(nsExpr ...string) string {
		return strings.Join(ParenAll(SkipEmpty(nsExpr)), " && ")
	}
	Neg = func(expr string) string {
		return fmt.Sprintf("!%s", expr)
	}
)

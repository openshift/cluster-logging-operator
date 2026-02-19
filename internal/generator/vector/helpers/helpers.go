package helpers

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	log "github.com/ViaQ/logerr/v2/log/static"
	v1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"

	"golang.org/x/sys/unix"
)

const (
	VectorSecretID = "kubernetes_secret"
	CLFSpec        = "clfSpec"
)

// Match quoted strings like "foo" or "foo/bar-baz"
var quoteRegex = regexp.MustCompile(`^".+"$`)

var (
	Replacer         = strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	listenAllAddress string
	listenAllOnce    sync.Once
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
	sort.Strings(out)
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
				if err = unix.Close(fd); err != nil {
					log.V(3).Error(err, "failed to close socket after determining local interface address")
				}
				return `[::]`
			}
		}()
	}
	listenAllOnce.Do(f)
	return listenAllAddress
}

// ConfigPath is the quoted path for any configmap visible to the collector
func ConfigPath(name string, file string, formatter ...string) string {
	formatString := "%q"
	if len(formatter) > 0 {
		formatString = formatter[0]
	}
	return fmt.Sprintf(formatString, filepath.Join(constants.ConfigMapBaseDir, name, file))
}

// SecretPath is the quoted path for any secret visible to the collector
func SecretPath(secretName string, file string, formatter ...string) string {
	formatString := "%q"
	if len(formatter) > 0 {
		formatString = formatter[0]
	}
	return fmt.Sprintf(formatString, filepath.Join(constants.CollectorSecretsDir, secretName, file))
}

// SecretFrom formated string SECRET[<secret_component_id>.<secret_name>#<secret_key>]
func SecretFrom(secretKey *v1.SecretReference) string {
	if secretKey != nil && secretKey.SecretName != "" && secretKey.Key != "" {
		return fmt.Sprintf("SECRET[%s.%s/%s]",
			VectorSecretID,
			secretKey.SecretName,
			secretKey.Key)
	}
	return ""
}

// GenerateQuotedPathSegmentArrayStr generates the final string of the array of array of path segments
// and array of flattened path with replaced not allowed symbols to feed into VRL
// E.g
// [.kubernetes.namespace_labels."bar/baz0-9.test"] -> ([["kubernetes","namespace_labels","bar/baz0-9.test"]], ["_kubernetes_namespace_labels_bar_baz0-9_test"])
func GenerateQuotedPathSegmentArrayStr(fieldPathArray []v1.FieldPath) (string, []string) {
	var quotedPathArray []string
	var flattenedArray []string

	for _, fieldPath := range fieldPathArray {
		pathStr := string(fieldPath)

		if strings.ContainsAny(pathStr, "/.") {
			flat := strings.NewReplacer(".", "_", "\"", "", "/", "_").Replace(pathStr)
			flat = strings.TrimPrefix(flat, "_")
			flattenedArray = append(flattenedArray, flat)
		}

		splitSegments := SplitPath(pathStr)
		quotedSegments := QuotePathSegments(splitSegments)
		quotedPathArray = append(quotedPathArray, fmt.Sprintf("[%s]", strings.Join(quotedSegments, ",")))
	}

	return fmt.Sprintf("[%s]", strings.Join(quotedPathArray, ",")),
		flattenedArray
}

// SplitPath splits a fieldPath by `.` and reassembles the quoted path segments that also contain `.`
// Example: `.foo."@some"."d.f.g.o111-22/333".foo_bar`
// Resultant Array: ["foo","@some",`"d.f.g.o111-22/333"`,"foo_bar"]
func SplitPath(path string) []string {
	var result []string

	segments := strings.Split(path, ".")

	var currSegment string
	for _, part := range segments {
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

// QuotePathSegments quotes all path segments as needed for VRL
func QuotePathSegments(pathArray []string) []string {
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

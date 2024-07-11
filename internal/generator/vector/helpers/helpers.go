package helpers

import (
	"fmt"
	v1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

const VectorSecretID = "kubernetes_secret"

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
				unix.Close(fd)
				return `[::]`
			}
		}()
	}
	listenAllOnce.Do(f)
	return listenAllAddress
}

// ConfigPath is the quoted path for any configmap visible to the collector
func ConfigPath(name string, file string) string {
	return fmt.Sprintf("%q", filepath.Join(constants.ConfigMapBaseDir, name, file))
}

// SecretPath is the quoted path for any secret visible to the collector
func SecretPath(secretName string, file string) string {
	return fmt.Sprintf("%q", filepath.Join(constants.CollectorSecretsDir, secretName, file))
}

// SecretFrom formated string SECRET[<secret_component_id>.<secret_name>#<secret_key>]
func SecretFrom(secretKey *v1.SecretConfigReference) string {
	if secretKey != nil && secretKey.Secret != nil && secretKey.Key != "" {
		return fmt.Sprintf("SECRET[%s.%s/%s]",
			VectorSecretID,
			secretKey.Secret.Name,
			secretKey.Key)
	}
	return ""
}

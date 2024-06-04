// package url provides operations on URLs and URL schemes for log forwarding destinations.
package url

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"net/url"
	"strings"
)

// CheckAbsolute returns an error if Scheme or Host is empty.
func CheckAbsolute(u *URL) error {
	switch {
	case u == nil:
		return fmt.Errorf("empty")
	case u.Scheme == "":
		return fmt.Errorf("no scheme: %q", u)
	case u.Host == "":
		return fmt.Errorf("no host: %q", u)
	}
	return nil
}

var secureSchemes = map[string]string{
	"https": "http",
	"udps":  "udp",
	"tls":   "tcp",
	"ssl":   "tcp",
}

// IsSecure determins if a URL specs a secure scheme
func IsSecure(urlString string) bool {
	if url, err := Parse(urlString); err == nil {
		return IsTLSScheme(url.Scheme)
	} else {
		log.V(0).Error(err, "Unable to parse URL", "url", urlString)
	}
	return true
}

// IsTLSScheme returns true if scheme is recognized as needing TLS security,
// for example "https", "tls"
func IsTLSScheme(scheme string) bool {
	return secureSchemes[strings.ToLower(scheme)] != ""
}

// PlainScheme returns the plain, non-TLS, version of scheme.
// Example: PlainScheme("https") == "http"
func PlainScheme(scheme string) string {
	scheme = strings.ToLower(scheme)
	if base, ok := secureSchemes[scheme]; ok {
		return base
	}
	return scheme
}

// Stubs for net/url types/functions so it's not necessary to import it as well.

type URL = url.URL

func Parse(s string) (*URL, error) { return url.Parse(s) }

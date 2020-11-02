// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package url

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseAbsolute parses an absolute URL - one with non-empty Scheme and Host.
// Returned errors include the string s.
func ParseAbsolute(s string) (*URL, error) {
	if s == "" {
		return nil, fmt.Errorf("empty")
	}
	u, err := url.Parse(s)
	switch {
	case err != nil:
		return nil, err
	case u.Scheme == "":
		return nil, fmt.Errorf("no scheme: %v", u)
	case u.Host == "":
		return nil, fmt.Errorf("no host: %v", u)
	}
	return u, nil
}

// ParseAbsoluteOrEmpty is like ParseAbsolute but returns an empty
// URL rather than an error if s == "".
func ParseAbsoluteOrEmpty(s string) (*URL, error) {
	if s == "" {
		return &url.URL{}, nil
	} else {
		return ParseAbsolute(s)
	}
}

// IsTLSScheme returns true if scheme is recognized as needing TLS security,
// for example "https", "tls" or "udps"
func IsTLSScheme(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "https", "tls", "udps":
		return true
	default:
		return false
	}
}

// Stubs for net/url types/functions so it's not necessary to import it as well.

type URL = url.URL

func Parse(s string) (*URL, error) { return url.Parse(s) }

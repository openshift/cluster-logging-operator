package consoleplugin

import (
	"golang.org/x/mod/semver"
	"strings"
)

const (
	featureDevConsole = "dev-console"
)

var (

	// featuresIfUnmatched represents the default features to enable
	featuresIfUnmatched = []string{featureDevConsole}

	versionMap = map[string][]string{
		"v4.10": {},
	}
)

// FeaturesForOCP will return the list of features to enable for the console plugin given the OCP version
func FeaturesForOCP(version string) []string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if features, found := versionMap[semver.MajorMinor(version)]; found {
		return features
	}
	return featuresIfUnmatched
}

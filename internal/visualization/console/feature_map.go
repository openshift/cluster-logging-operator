package console

import (
	"golang.org/x/mod/semver"
	"strings"
)

const (
	featureDevConsole = "dev-console"
	featureAlerts     = "alerts"
)

var (

	// featuresIfUnmatched represents the default features to enable
	featuresIfUnmatched = []string{featureAlerts, featureDevConsole}
)

// FeaturesForOCP will return the list of features to enable for the console plugin given the OCP version
func FeaturesForOCP(version string) []string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if semver.Compare(version, "v4.11") < 0 {
		return []string{}
	}
	if semver.Compare(version, "v4.11") >= 0 && semver.Compare(version, "v4.13") < 0 {
		return []string{featureDevConsole}
	}
	return featuresIfUnmatched
}

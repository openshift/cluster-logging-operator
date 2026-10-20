package tls

import (
	"os"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
)

// fipsEnabledFile is the kernel interface reporting whether the host is
// running in FIPS mode. FIPS mode is an install-time setting that applies to
// every node of the cluster, so the value read by the operator pod reflects
// the whole cluster. It is a variable so tests can point it at a fixture.
var fipsEnabledFile = "/proc/sys/crypto/fips_enabled"

// nonFIPSTLSGroups are the post-quantum hybrid TLS groups, which are not
// FIPS-approved and must be omitted by components running in FIPS mode:
//
//	"X25519MLKEM768 is a post-quantum hybrid group that is not FIPS-approved
//	 and should be ignored by components running in FIPS mode."
//
// (openshift/api configv1.TLSGroup). Vector passes the configured groups to
// OpenSSL's SSL_CTX_set1_groups_list, which fails — and the collector
// crash-loops — when any group in the list is unavailable to the active
// providers, and the FIPS provider does not provide these groups.
var nonFIPSTLSGroups = map[configv1.TLSGroup]struct{}{
	configv1.TLSGroupX25519MLKEM768:     {},
	configv1.TLSGroupSecP256r1MLKEM768:  {},
	configv1.TLSGroupSecP384r1MLKEM1024: {},
}

// FIPSEnabled reports whether the operator's host is running in FIPS mode.
// A missing or unreadable flag is treated as "not FIPS" so that detection
// failures preserve the historical behavior rather than guess.
func FIPSEnabled() bool {
	enabled, err := os.ReadFile(fipsEnabledFile)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(enabled)) == "1"
}

// FilterGroupsForFIPS removes the post-quantum hybrid groups, which are not
// FIPS-approved, when the host is running in FIPS mode. When FIPS is not
// enabled the groups are returned unchanged.
func FilterGroupsForFIPS(groups []configv1.TLSGroup) []configv1.TLSGroup {
	if !FIPSEnabled() {
		return groups
	}
	filtered := make([]configv1.TLSGroup, 0, len(groups))
	for _, group := range groups {
		if _, drop := nonFIPSTLSGroups[group]; drop {
			log.V(1).Info("Dropping TLS group that is not FIPS-approved", "group", group)
			continue
		}
		filtered = append(filtered, group)
	}
	return filtered
}

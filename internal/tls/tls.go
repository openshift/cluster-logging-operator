// Package tls provides TLS profile management for cluster logging components.
// It enforces CWE-327 security hardening: deprecated TLS versions (1.0, 1.1) are rejected,
// and only secure AEAD ciphers (AES-GCM, ChaCha20-Poly1305) with ECDHE key exchange are accepted.
// OpenSSL-named ciphers from OpenShift TLS profiles are validated against a secure mapping.
package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	APIServerName = "cluster"
)

var (
	// DefaultTLSProfileType is the intermediate profile type
	DefaultTLSProfileType = configv1.TLSProfileIntermediateType
	// DefaultTLSCiphers are the default TLS ciphers for API servers
	DefaultTLSCiphers = configv1.TLSProfiles[DefaultTLSProfileType].Ciphers
	// DefaultMinTLSVersion is the default minimum TLS version for API servers
	DefaultMinTLSVersion = configv1.TLSProfiles[DefaultTLSProfileType].MinTLSVersion
	// DefaultTLSGroups are the default TLS groups (elliptic curves) for key exchange.
	DefaultTLSGroups = tlsGroupsToStrings(configv1.TLSProfiles[DefaultTLSProfileType].Groups)

	// supportedTLSGroups maps OpenShift API TLS group names to Go tls.CurveID values.
	// Go 1.25 only provides tls.X25519MLKEM768; it lacks CurveID constants for
	// TLSGroupSecP256r1MLKEM768 and TLSGroupSecP384r1MLKEM1024. Those groups are
	// still forwarded to Vector via opensslGroupNames (OpenSSL supports them), but
	// TLSConfigFromProfile will skip them when building the Go tls.Config.
	supportedTLSGroups = map[configv1.TLSGroup]tls.CurveID{
		configv1.TLSGroupX25519:         tls.X25519,
		configv1.TLSGroupSecP256r1:      tls.CurveP256,
		configv1.TLSGroupSecP384r1:      tls.CurveP384,
		configv1.TLSGroupSecP521r1:      tls.CurveP521,
		configv1.TLSGroupX25519MLKEM768: tls.X25519MLKEM768,
	}

	// opensslGroupNames maps OpenShift API TLS group names to OpenSSL curve names.
	// Most names are the same; secp256r1 is the notable exception.
	opensslGroupNames = map[configv1.TLSGroup]string{
		configv1.TLSGroupX25519:             "X25519",
		configv1.TLSGroupSecP256r1:          "prime256v1",
		configv1.TLSGroupSecP384r1:          "secp384r1",
		configv1.TLSGroupSecP521r1:          "secp521r1",
		configv1.TLSGroupX25519MLKEM768:     "X25519MLKEM768",
		configv1.TLSGroupSecP256r1MLKEM768:  "SecP256r1MLKEM768",
		configv1.TLSGroupSecP384r1MLKEM1024: "SecP384r1MLKEM1024",
	}

	// openSSLToIANACiphersMap maps OpenSSL cipher suite names (used in OpenShift TLS profiles)
	// to their IANA equivalents (used by Go's crypto/tls). Only secure AEAD ciphers with
	// ECDHE key exchange are included (CWE-327 hardening). This enforces the same
	// restrictions as the log-file-metric-exporter binary, preventing the operator from
	// passing insecure ciphers (CBC, 3DES, SHA-1 MACs, non-ECDHE) to pod TLS configuration.
	// See LOG-9764 for the security rationale.
	openSSLToIANACiphersMap = map[string]string{
		"ECDHE-ECDSA-AES128-GCM-SHA256":  "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"ECDHE-RSA-AES128-GCM-SHA256":    "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"ECDHE-ECDSA-AES256-GCM-SHA384":  "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"ECDHE-RSA-AES256-GCM-SHA384":    "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"ECDHE-ECDSA-CHACHA20-POLY1305": "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
		"ECDHE-RSA-CHACHA20-POLY1305":   "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
	}
)

// FetchAPIServerTlsProfile fetches tlsSecurityProfile configured in APIServer
func FetchAPIServerTlsProfile(k8client client.Client) (*configv1.TLSSecurityProfile, error) {
	apiServer := &configv1.APIServer{}
	key := client.ObjectKey{Name: APIServerName}
	if err := k8client.Get(context.TODO(), key, apiServer); err != nil {
		return nil, err
	}
	return apiServer.Spec.TLSSecurityProfile, nil
}

// TLSCiphers returns the TLS ciphers for the TLS security profile.
// Insecure cipher suites are filtered out (CWE-327): only AEAD ciphers (AES-GCM, ChaCha20-Poly1305)
// with ECDHE key exchange are accepted. OpenSSL-named ciphers from OpenShift profiles are validated
// against openSSLToIANACiphersMap; ciphers not in the mapping (CBC, 3DES, SHA-1 MACs, non-ECDHE)
// are silently dropped. If all ciphers are filtered, defaults to the Intermediate profile.
// The returned cipher list can be passed to the exporter binary via -cipherSuites flag (comma-separated).
func TLSCiphers(profile configv1.TLSProfileSpec) []string {
	if len(profile.Ciphers) == 0 {
		return DefaultTLSCiphers
	}

	// Build a map of secure IANA cipher names for fast lookup
	secureCipherNames := make(map[string]bool)
	for _, suite := range tls.CipherSuites() {
		secureCipherNames[suite.Name] = true
	}

	// Build a reverse map of IANA names to OpenSSL names for checking if an OpenSSL cipher is secure
	secureCiphersByOpenSSL := make(map[string]bool)
	for openSSLName, ianaName := range openSSLToIANACiphersMap {
		if secureCipherNames[ianaName] {
			secureCiphersByOpenSSL[openSSLName] = true
		}
	}

	// TLS 1.3 cipher suites (not in tls.CipherSuites() but still secure)
	// Explicit allowlist to reject CCM variants
	tls13Ciphers := map[string]bool{
		"TLS_AES_128_GCM_SHA256":       true,
		"TLS_AES_256_GCM_SHA384":       true,
		"TLS_CHACHA20_POLY1305_SHA256": true,
	}

	validCiphers := make([]string, 0, len(profile.Ciphers))
	for _, cipherName := range profile.Ciphers {
		// Accept if it's in the secure IANA cipher list
		if secureCipherNames[cipherName] {
			validCiphers = append(validCiphers, cipherName)
			continue
		}

		// Check if this is a secure TLS 1.3 cipher
		if tls13Ciphers[cipherName] {
			validCiphers = append(validCiphers, cipherName)
			continue
		}

		// Check if this is a secure OpenSSL-named cipher
		if secureCiphersByOpenSSL[cipherName] {
			validCiphers = append(validCiphers, cipherName)
			continue
		}

		log.V(1).Info("Filtering out insecure or unsupported cipher suite", "cipher", cipherName)
	}

	// If all ciphers were filtered out, fall back to defaults
	if len(validCiphers) == 0 {
		log.Info("All ciphers were filtered out as insecure, falling back to defaults",
			"originalCount", len(profile.Ciphers),
			"defaultCount", len(DefaultTLSCiphers))
		return DefaultTLSCiphers
	}

	return validCiphers
}

// MinTLSVersion returns the minimum TLS version for the TLS security profile.
// Deprecated TLS versions (1.0, 1.1) are rejected and automatically upgraded to TLS 1.2 (CWE-327).
// This ensures components deployed with the Old TLS profile receive a secure minimum version.
// The returned version string can be passed to the exporter binary via -tlsMinVersion flag.
func MinTLSVersion(profile configv1.TLSProfileSpec) string {
	if profile.MinTLSVersion == "" {
		return string(DefaultMinTLSVersion)
	}
	// Reject deprecated TLS versions - return TLS 1.2 as safe fallback
	if profile.MinTLSVersion == configv1.VersionTLS10 || profile.MinTLSVersion == configv1.VersionTLS11 {
		log.Info("Rejecting deprecated TLS version, using TLS 1.2 instead",
			"requestedVersion", profile.MinTLSVersion,
			"fallbackVersion", configv1.VersionTLS12)
		return string(configv1.VersionTLS12)
	}
	return string(profile.MinTLSVersion)
}

// TLSGroups returns the TLS groups for the TLS security profile.
func TLSGroups(profile configv1.TLSProfileSpec) []string {
	if len(profile.Groups) == 0 {
		return DefaultTLSGroups
	}
	return tlsGroupsToStrings(profile.Groups)
}

// TLSGroupToID converts an OpenShift API TLS group name to a Go crypto/tls CurveID.
func TLSGroupToID(group configv1.TLSGroup) (tls.CurveID, error) {
	if id, ok := supportedTLSGroups[group]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("unsupported TLS group: %s", group)
}

// TLSGroupsToOpenSSL converts OpenShift API TLS groups to a colon-separated
// string of OpenSSL curve names for use in Vector TLS configuration.
func TLSGroupsToOpenSSL(groups []configv1.TLSGroup) string {
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		if name, ok := opensslGroupNames[g]; ok {
			names = append(names, name)
		} else {
			log.V(1).Info("Skipping unknown TLS group for OpenSSL", "group", g)
		}
	}
	return strings.Join(names, ":")
}

func tlsGroupsToStrings(groups []configv1.TLSGroup) []string {
	s := make([]string, len(groups))
	for i, g := range groups {
		s[i] = string(g)
	}
	return s
}

// GetClusterTLSProfileSpec returns TLSProfileSpec
func GetClusterTLSProfileSpec(apiServerTLSProfile *configv1.TLSSecurityProfile) configv1.TLSProfileSpec {
	defaultProfile := *configv1.TLSProfiles[DefaultTLSProfileType]
	if apiServerTLSProfile == nil || apiServerTLSProfile.Type == "" {
		return defaultProfile
	}
	profileType := apiServerTLSProfile.Type

	if profileType != configv1.TLSProfileCustomType {
		if tlsConfig, ok := configv1.TLSProfiles[profileType]; ok {
			return *tlsConfig
		}
		return defaultProfile
	}

	if apiServerTLSProfile.Custom != nil {
		return apiServerTLSProfile.Custom.TLSProfileSpec
	}

	return defaultProfile
}

// CipherSuiteStringToID converts cipher suite name to crypto/tls ID.
// Only secure cipher suites are supported; insecure ciphers are rejected (CWE-327).
func CipherSuiteStringToID(name string) (uint16, error) {
	for _, suite := range tls.CipherSuites() {
		if suite.Name == name {
			return suite.ID, nil
		}
	}
	return 0, fmt.Errorf("unsupported cipher suite: %s", name)
}

// TLSVersionToConstant converts TLS version string to crypto/tls constant.
// Only TLS 1.2 and 1.3 are supported; deprecated versions are rejected (CWE-327).
func TLSVersionToConstant(version configv1.TLSProtocolVersion) (uint16, error) {
	switch version {
	case configv1.VersionTLS12:
		return tls.VersionTLS12, nil
	case configv1.VersionTLS13:
		return tls.VersionTLS13, nil
	default:
		return tls.VersionTLS12, nil // Default to TLS 1.2
	}
}

// TLSConfigFromProfile creates a crypto/tls.Config from TLSProfileSpec
func TLSConfigFromProfile(profileSpec configv1.TLSProfileSpec) (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12, // Safe default
	}

	if profileSpec.MinTLSVersion != "" {
		minVersion, err := TLSVersionToConstant(profileSpec.MinTLSVersion)
		if err != nil {
			return nil, err
		}
		config.MinVersion = minVersion
	}

	if len(profileSpec.Ciphers) > 0 {
		cipherSuites := make([]uint16, 0, len(profileSpec.Ciphers))
		for _, cipherName := range profileSpec.Ciphers {
			id, err := CipherSuiteStringToID(cipherName)
			if err != nil {
				log.V(1).Info("Skipping unsupported cipher suite", "cipher", cipherName)
				continue
			}
			cipherSuites = append(cipherSuites, id)
		}
		if len(cipherSuites) > 0 {
			config.CipherSuites = cipherSuites
		}
	}

	if len(profileSpec.Groups) > 0 {
		curvePreferences := make([]tls.CurveID, 0, len(profileSpec.Groups))
		for _, group := range profileSpec.Groups {
			curveID, err := TLSGroupToID(group)
			if err != nil {
				log.V(1).Info("Skipping unsupported TLS group", "group", group)
				continue
			}
			curvePreferences = append(curvePreferences, curveID)
		}
		if len(curvePreferences) > 0 {
			config.CurvePreferences = curvePreferences
		}
	}

	return config, nil
}

// GetTLSConfigOptions returns TLS config options for the manager
func GetTLSConfigOptions(k8sClient client.Client) ([]func(*tls.Config), error) {
	tlsProfile, err := FetchAPIServerTlsProfile(k8sClient)
	if err != nil {
		log.V(1).Info("Failed to fetch APIServer TLS profile, using defaults", "error", err)
		tlsProfile = nil
	}

	profileSpec := GetClusterTLSProfileSpec(tlsProfile)
	tlsConfig, err := TLSConfigFromProfile(profileSpec)
	if err != nil {
		return nil, err
	}

	log.Info("Configured TLS profile", "minVersion", tlsConfig.MinVersion, "cipherSuites", len(tlsConfig.CipherSuites))

	return []func(*tls.Config){
		func(cfg *tls.Config) {
			cfg.MinVersion = tlsConfig.MinVersion
			cfg.CipherSuites = tlsConfig.CipherSuites
			cfg.CurvePreferences = tlsConfig.CurvePreferences
		},
	}, nil
}

// IsClusterAPIServer returns true if the object is the cluster APIServer resource
func IsClusterAPIServer(obj client.Object) bool {
	apiServer, ok := obj.(*configv1.APIServer)
	return ok && apiServer.Name == APIServerName
}

// APIServerTLSProfileChangedPredicate returns a predicate that filters APIServer events
// to only those that involve changes to the TLS security profile.
// The reconcileOnCreate parameter controls whether to reconcile when the APIServer is created.
func APIServerTLSProfileChangedPredicate(reconcileOnCreate bool) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if !reconcileOnCreate {
				return false
			}
			return IsClusterAPIServer(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if !IsClusterAPIServer(e.ObjectNew) {
				return false
			}
			oldAPIServer, oldOk := e.ObjectOld.(*configv1.APIServer)
			newAPIServer, newOk := e.ObjectNew.(*configv1.APIServer)
			if !oldOk || !newOk {
				return false
			}
			// Only trigger if the TLS profile has changed
			return !reflect.DeepEqual(oldAPIServer.Spec.TLSSecurityProfile, newAPIServer.Spec.TLSSecurityProfile)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false // Don't reconcile on delete
		},
	}
}

package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"

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
	// TODO: When openshift/api is updated to include Groups in TLSProfileSpec,
	// read from configv1.TLSProfiles[DefaultTLSProfileType].Groups directly.
	DefaultTLSGroups = []string{"X25519", "secp256r1", "secp384r1", "X25519MLKEM768"}
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

// TLSCiphers returns the TLS ciphers for the
// TLS security profile defined in the APIServerConfig.
func TLSCiphers(profile configv1.TLSProfileSpec) []string {
	if len(profile.Ciphers) == 0 {
		return DefaultTLSCiphers
	}
	return profile.Ciphers
}

// MinTLSVersion returns the minimum TLS version for the
// TLS security profile defined in the APIServerConfig.
func MinTLSVersion(profile configv1.TLSProfileSpec) string {
	if profile.MinTLSVersion == "" {
		return string(DefaultMinTLSVersion)
	}
	return string(profile.MinTLSVersion)
}

// TLSGroups returns the TLS groups for key exchange.
// TODO: When openshift/api is updated with the Groups field in TLSProfileSpec,
// update this to accept configv1.TLSProfileSpec like TLSCiphers and MinTLSVersion.
func TLSGroups(groups []string) []string {
	if len(groups) == 0 {
		return DefaultTLSGroups
	}
	return groups
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

// CipherSuiteStringToID converts cipher suite name to crypto/tls ID
func CipherSuiteStringToID(name string) (uint16, error) {
	for _, suite := range tls.CipherSuites() {
		if suite.Name == name {
			return suite.ID, nil
		}
	}
	for _, suite := range tls.InsecureCipherSuites() {
		if suite.Name == name {
			return suite.ID, nil
		}
	}
	return 0, fmt.Errorf("unsupported cipher suite: %s", name)
}

// TLSVersionToConstant converts TLS version string to crypto/tls constant
func TLSVersionToConstant(version configv1.TLSProtocolVersion) (uint16, error) {
	switch version {
	case configv1.VersionTLS10:
		return tls.VersionTLS10, nil
	case configv1.VersionTLS11:
		return tls.VersionTLS11, nil
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

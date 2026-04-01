package tls

import (
	"context"
	"reflect"

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
			return !reflect.DeepEqual(oldAPIServer.Spec.TLSSecurityProfile, newAPIServer.Spec.TLSSecurityProfile)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

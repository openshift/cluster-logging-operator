package clusterlogforwarder

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// validateUrlAccordingToTls validate that if Output has TLS configuration Output URL scheme must be secure e.g. https, tls etc
func validateUrlAccordingToTls(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus) {
	for i, output := range clf.Spec.Outputs {
		_, output := i, output // Don't bind range variable.
		u, _ := url.Parse(output.URL)
		scheme := strings.ToLower(u.Scheme)
		if !url.IsTLSScheme(scheme) && (output.TLS != nil && (output.TLS.InsecureSkipVerify || output.TLS.TLSSecurityProfile != nil)) {
			log.V(3).Info("validateUrlAccordingToTls failed", "reason", "URL not secure but output has TLS configuration parameters",
				"output URL", output.URL, "output Name", output.Name)
			return fmt.Errorf("URL not secure: %v, but output %s has TLS configuration parameters", u, output.Name), nil
		}
	}
	return nil, nil
}

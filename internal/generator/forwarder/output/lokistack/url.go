package lokistack

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
)

const (
	// OtlpEndpoint is the OTLP logs endpoint path for LokiStack
	OtlpEndpoint = "/otlp/v1/logs"
)

// GatewayService returns the service name for the LokiStack gateway.
func GatewayService(lokiStackName string) string {
	return fmt.Sprintf("%s-gateway-http", lokiStackName)
}

// URL constructs the base URL for a LokiStack tenant.
// Format: https://<gateway-service>.<namespace>.svc:8080/api/logs/v1/<tenant>
func URL(ls *obs.LokiStack, tenant string) string {
	if !observability.ReservedInputTypes.Has(tenant) {
		return ""
	}

	service := GatewayService(ls.Target.Name)
	return fmt.Sprintf("https://%s.%s.svc:8080/api/logs/v1/%s", service, ls.Target.Namespace, tenant)
}

// OtlpURL constructs the OTLP endpoint URL for a LokiStack tenant.
// Format: https://<gateway-service>.<namespace>.svc:8080/api/logs/v1/<tenant>/otlp/v1/logs
func OtlpURL(ls *obs.LokiStack, tenant string) string {
	baseURL := URL(ls, tenant)
	if baseURL == "" {
		return ""
	}
	return baseURL + OtlpEndpoint
}

package opentelemetry

import (
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Initializer is a function that knows how to initialize a kubernetes runtime object
type Initializer func(o runtime.Object, namespace, name string, visitors ...func(o runtime.Object))

// NewOpenTelemetryCollector returns an OpenTelemetryCollector with name and namespace.
func NewOpenTelemetryCollector(namespace, name string, initialize Initializer, visitors ...func(otelcol *otelv1beta1.OpenTelemetryCollector)) *otelv1beta1.OpenTelemetryCollector {
	otelcol := &otelv1beta1.OpenTelemetryCollector{}
	initialize(otelcol, namespace, name)
	for _, v := range visitors {
		v(otelcol)
	}
	return otelcol
}

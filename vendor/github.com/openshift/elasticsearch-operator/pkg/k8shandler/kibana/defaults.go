package kibana

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultKibanaMemory     resource.Quantity = resource.MustParse("736Mi")
	defaultKibanaCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultKibanaProxyMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultKibanaProxyCpuRequest resource.Quantity = resource.MustParse("100m")
)

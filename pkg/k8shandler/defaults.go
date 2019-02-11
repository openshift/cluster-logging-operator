package k8shandler

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultEsMemory     resource.Quantity = resource.MustParse("16Gi")
	defaultEsCpuRequest resource.Quantity = resource.MustParse("1000m")

	defaultKibanaMemory     resource.Quantity = resource.MustParse("736Mi")
	defaultKibanaCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultKibanaProxyMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultKibanaProxyCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultCuratorMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultCuratorCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultFluentdMemory     resource.Quantity = resource.MustParse("736Mi")
	defaultFluentdCpuRequest resource.Quantity = resource.MustParse("100m")

	defaultRsyslogMemory     resource.Quantity = resource.MustParse("358Mi")
	defaultRsyslogCpuRequest resource.Quantity = resource.MustParse("100m")
)

package k8shandler

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultEsMemory     resource.Quantity = resource.MustParse("16Gi")
	defaultEsCpuRequest resource.Quantity = resource.MustParse("1")

	defaultEsProxyMemory     resource.Quantity = resource.MustParse("256Mi")
	defaultEsProxyCpuRequest resource.Quantity = resource.MustParse("100m")
)

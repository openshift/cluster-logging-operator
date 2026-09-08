package lokistack

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	lokiGVR = schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}
)

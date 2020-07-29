package k8s

import "k8s.io/apimachinery/pkg/runtime"

type Client interface {
	Create(object runtime.Object) error

	Get(objectName string, object runtime.Object) error

	Delete(object runtime.Object) error

	Update(object runtime.Object) (err error)
}

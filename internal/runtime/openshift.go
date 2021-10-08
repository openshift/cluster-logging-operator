package runtime

import (
	openshiftv1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	must(openshiftv1.AddToScheme(scheme.Scheme))
}

// NewRoute returns an openshift.io/v1.Route with namespace and name to port on service.
func NewRoute(namespace, name, service, port string) *openshiftv1.Route {
	r := &openshiftv1.Route{}
	Initialize(r, namespace, name)
	r.Spec.Port = &openshiftv1.RoutePort{TargetPort: intstr.Parse(port)}
	r.Spec.To.Kind = "Service"
	r.Spec.To.Name = service
	return r
}

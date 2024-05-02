package k8shandler

import (
	"fmt"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	route "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRoute stubs an instance of a Route
func NewRoute(routeName, namespace, serviceName, cafilePath string) *route.Route {
	return &route.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: route.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Namespace: namespace,
			Labels: map[string]string{
				"component":     "support",
				"logging-infra": "support",
				"provider":      "openshift",
			},
		},
		Spec: route.RouteSpec{
			To: route.RouteTargetReference{
				Name: serviceName,
				Kind: "Service",
			},
			TLS: &route.TLSConfig{
				Termination:                   route.TLSTerminationReencrypt,
				InsecureEdgeTerminationPolicy: route.InsecureEdgeTerminationPolicyRedirect,
				CACertificate:                 string(utils.GetFileContents(cafilePath)),
				DestinationCACertificate:      string(utils.GetFileContents(cafilePath)),
			},
		},
	}
}

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateRoute(newRoute *route.Route) error {

	err := clusterRequest.Create(newRoute)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating route for %q: %v", clusterRequest.Cluster.Name, err)
		}

		// else -- try to update it if its a valid change (e.g. spec.tls)
		current := &route.Route{}

		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := clusterRequest.Get(newRoute.Name, current); err != nil {
				return fmt.Errorf("Failed to get route: %v", err)
			}

			if !reflect.DeepEqual(current.Spec.TLS, newRoute.Spec.TLS) {
				current.Spec.TLS = newRoute.Spec.TLS
				return clusterRequest.Update(current)
			}

			return nil
		})
	}

	return nil
}

// GetRouteURL retrieves the route URL from a given route and namespace
func (clusterRequest *ClusterLoggingRequest) GetRouteURL(routeName string) (string, error) {

	foundRoute := &route.Route{}

	if err := clusterRequest.Get(routeName, foundRoute); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "Failed to check for ClusterLogging object")
		}
		return "", err
	}

	return fmt.Sprintf("%s%s", "https://", foundRoute.Spec.Host), nil
}

// RemoveRoute with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveRoute(routeName string) error {

	rt := NewRoute(
		routeName,
		clusterRequest.Cluster.Namespace,
		routeName,
		"",
	)

	err := clusterRequest.Delete(rt)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v route %v", routeName, err)
	}

	return nil
}

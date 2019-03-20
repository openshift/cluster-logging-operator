package utils

import (
	"fmt"

	route "github.com/openshift/api/route/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewRoute stubs an instance of a Route
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
				CACertificate:                 string(GetFileContents(cafilePath)),
				DestinationCACertificate:      string(GetFileContents(cafilePath)),
			},
		},
	}
}

//GetRouteURL retrieves the route URL from a given route and namespace
func GetRouteURL(routeName, namespace string) (string, error) {

	foundRoute := &route.Route{
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
	}

	if err := sdk.Get(foundRoute); err != nil {
		if !errors.IsNotFound(err) {
			logrus.Errorf("Failed to check for ClusterLogging object: %v", err)
		}
		return "", err
	}

	return fmt.Sprintf("%s%s", "https://", foundRoute.Spec.Host), nil
}

//RemoveRoute with given name and namespace
func RemoveRoute(namespace, routeName string) error {

	route := NewRoute(
		routeName,
		namespace,
		routeName,
		"",
	)

	err := sdk.Delete(route)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v route %v", routeName, err)
	}

	return nil
}

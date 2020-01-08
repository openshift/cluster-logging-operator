package kibana

import (
	"fmt"
	"strings"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"

	route "github.com/openshift/api/route/v1"
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
				CACertificate:                 string(utils.GetFileContents(cafilePath)),
				DestinationCACertificate:      string(utils.GetFileContents(cafilePath)),
			},
		},
	}
}

//GetRouteURL retrieves the route URL from a given route and namespace
func (clusterRequest *KibanaRequest) GetRouteURL(routeName string) (string, error) {

	foundRoute := &route.Route{}

	if err := clusterRequest.Get(routeName, foundRoute); err != nil {
		if !errors.IsNotFound(err) {
			logrus.Errorf("Failed to check for kibana object: %v", err)
		}
		return "", err
	}

	return fmt.Sprintf("%s%s", "https://", foundRoute.Spec.Host), nil
}

//RemoveRoute with given name and namespace
func (clusterRequest *KibanaRequest) RemoveRoute(routeName string) error {

	route := NewRoute(
		routeName,
		clusterRequest.cluster.Namespace,
		routeName,
		"",
	)

	err := clusterRequest.Delete(route)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v route %v", routeName, err)
	}

	return nil
}

func (clusterRequest *KibanaRequest) createOrUpdateKibanaRoute() error {

	cluster := clusterRequest.cluster

	kibanaRoute := NewRoute(
		"kibana",
		cluster.Namespace,
		"Kibana",
		utils.GetWorkingDirFilePath("ca.crt"),
	)

	utils.AddOwnerRefToObject(kibanaRoute, utils.AsOwner(cluster))

	err := clusterRequest.Create(kibanaRoute)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route for %q: %v", cluster.Name, err)
	}

	kibanaURL, err := clusterRequest.GetRouteURL("kibana")
	if err != nil {
		return err
	}

	sharedConfig := createSharedConfig(cluster.Namespace, kibanaURL, kibanaURL)
	utils.AddOwnerRefToObject(sharedConfig, utils.AsOwner(cluster))

	err = clusterRequest.Create(sharedConfig)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config: %v", err)
	}

	sharedRole := NewRole(
		"sharing-config-reader",
		cluster.Namespace,
		NewPolicyRules(
			NewPolicyRule(
				[]string{""},
				[]string{"configmaps"},
				[]string{"sharing-config"},
				[]string{"get"},
			),
		),
	)

	utils.AddOwnerRefToObject(sharedRole, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(sharedRole)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config role for %q: %v", cluster.Name, err)
	}

	sharedRoleBinding := NewRoleBinding(
		"openshift-logging-sharing-config-reader-binding",
		cluster.Namespace,
		"sharing-config-reader",
		NewSubjects(
			NewSubject(
				"Group",
				"system:authenticated",
			),
		),
	)

	utils.AddOwnerRefToObject(sharedRoleBinding, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(sharedRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config role binding for %q: %v", cluster.Name, err)
	}

	return nil
}

func (clusterRequest *KibanaRequest) createOrUpdateKibanaConsoleExternalLogLink() (err error) {
	cluster := clusterRequest.cluster

	kibanaURL, err := clusterRequest.GetRouteURL("kibana")
	if err != nil {
		return err
	}

	consoleExternalLogLink := NewConsoleExternalLogLink(
		"kibana",
		cluster.Namespace,
		"Show in Kibana",
		strings.Join([]string{kibanaURL,
			"/app/kibana#/discover?_g=(time:(from:now-1w,mode:relative,to:now))&_a=(columns:!(kubernetes.container_name,message),query:(query_string:(analyze_wildcard:!t,query:'",
			strings.Join([]string{
				"kubernetes.pod_name:\"${resourceName}\"",
				"kubernetes.namespace_name:\"${resourceNamespace}\"",
				"kubernetes.container_name.raw:\"${containerName}\"",
			}, " AND "),
			"')),sort:!('@timestamp',desc))"},
			""),
	)

	utils.AddOwnerRefToObject(consoleExternalLogLink, utils.AsOwner(cluster))

	// In case the object already exists we delete it first
	if err = clusterRequest.RemoveConsoleExternalLogLink("kibana"); err != nil {
		return
	}

	err = clusterRequest.Create(consoleExternalLogLink)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana console external log link for %q: %v", cluster.Name, err)
	}
	return nil
}

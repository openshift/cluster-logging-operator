package k8shandler

import (
	"fmt"
	"reflect"
	"sort"

	"k8s.io/apimachinery/pkg/api/errors"

	oauth "github.com/openshift/api/oauth/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewOAuthClient stubs a new OAuthClient
func NewOAuthClient(oauthClientName, namespace, oauthSecret string, redirectURIs, scopeRestrictions []string) *oauth.OAuthClient {

	return &oauth.OAuthClient{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuthClient",
			APIVersion: oauth.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: oauthClientName,
			Labels: map[string]string{
				"logging-infra": "support",
				"namespace":     namespace,
			},
		},
		Secret:       oauthSecret,
		RedirectURIs: redirectURIs,
		GrantMethod:  oauth.GrantHandlerPrompt,
		ScopeRestrictions: []oauth.ScopeRestriction{
			{
				ExactValues: scopeRestrictions,
			},
		},
	}

}

//RemoveOAuthClient with a given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveOAuthClient(clientName string) error {

	oauthClient := NewOAuthClient(
		clientName,
		clusterRequest.Cluster.Namespace,
		"",
		[]string{},
		[]string{},
	)

	//TODO: Remove this in the next release after removing old kibana code completely
	if !HasCLORef(oauthClient, clusterRequest) {
		return nil
	}

	err := clusterRequest.Delete(oauthClient)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v oauthclient %v", clientName, err)
	}

	return nil
}

//areOAuthClientsSame compares two OAuthClients for sameness defined as: redirectURIs, secret, scopeRestrictions
//this function makes no attempt to normalize the scopeRestrictions
func areOAuthClientsSame(first, second *oauth.OAuthClient) bool {
	sort.Strings(first.RedirectURIs)
	sort.Strings(second.RedirectURIs)
	return first.Secret == second.Secret && reflect.DeepEqual(first.RedirectURIs, second.RedirectURIs) && reflect.DeepEqual(first.ScopeRestrictions, second.ScopeRestrictions)
}

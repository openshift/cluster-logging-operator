package utils

import (
	"fmt"

	oauth "github.com/openshift/api/oauth/v1"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/api/errors"
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
			oauth.ScopeRestriction{
				ExactValues: scopeRestrictions,
			},
		},
	}

}

//RemoveOAuthClient with a given name and namespace
func RemoveOAuthClient(namespace, clientName string) error {

	oauthClient := NewOAuthClient(
		clientName,
		namespace,
		"",
		[]string{},
		[]string{},
	)

	err := sdk.Delete(oauthClient)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v oauthclient %v", clientName, err)
	}

	return nil
}

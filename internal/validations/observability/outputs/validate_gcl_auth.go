package outputs

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

const (
	gclCredTypeServiceAccount  = "service_account"
	gclCredTypeExternalAccount = "external_account"
)

type gcpCredentialFile struct {
	Type             string               `json:"type"`
	CredentialSource *gcpCredentialSource `json:"credential_source"`
}

type gcpCredentialSource struct {
	File string `json:"file"`
}

func ValidateGCLAuth(o obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	if o.GoogleCloudLogging == nil || o.GoogleCloudLogging.Authentication == nil {
		return results
	}
	auth := o.GoogleCloudLogging.Authentication
	secrets := observability.Secrets(context.Secrets)

	if auth.Credentials == nil {
		return append(results, "GCP authentication is missing credentials")
	}

	creds, err := parseGCPCredentials(auth.Credentials, secrets)
	if err != nil {
		return append(results, err.Error())
	}

	switch creds.Type {
	case gclCredTypeServiceAccount:
		// Service account key — no additional validation needed
	case gclCredTypeExternalAccount:
		results = append(results, validateGCLExternalAccount(creds, auth.Token)...)
	default:
		results = append(results, fmt.Sprintf("GCP credentials file has unsupported type %q, expected %q or %q", creds.Type, gclCredTypeServiceAccount, gclCredTypeExternalAccount))
	}

	return results
}

func validateGCLExternalAccount(creds *gcpCredentialFile, token *obs.BearerToken) (results []string) {
	if creds.CredentialSource == nil {
		return append(results, "GCP external account credentials missing required field \"credential_source\"")
	}

	expectedPath := expectedTokenPath(token)
	if expectedPath != "" && creds.CredentialSource.File != expectedPath {
		results = append(results, fmt.Sprintf("GCP external account credential_source.file %q does not match expected token path %q", creds.CredentialSource.File, expectedPath))
	}

	return results
}

func parseGCPCredentials(ref *obs.SecretReference, secrets observability.Secrets) (*gcpCredentialFile, error) {
	data := secrets.Value(ref)
	if data == nil {
		return nil, fmt.Errorf("GCP credentials secret %q with key %q not found", ref.SecretName, ref.Key)
	}

	var creds gcpCredentialFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("GCP credentials secret is not valid JSON")
	}
	return &creds, nil
}

func expectedTokenPath(token *obs.BearerToken) string {
	if token == nil {
		return ""
	}
	switch token.From {
	case obs.BearerTokenFromServiceAccount:
		return filepath.Join(constants.ServiceAccountSecretPath, constants.TokenKey)
	case obs.BearerTokenFromSecret:
		if token.Secret != nil {
			return filepath.Join(constants.CollectorSecretsDir, token.Secret.Name, token.Secret.Key)
		}
	}
	return ""
}

package cloudwatch

import (
	_ "embed"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/version"
	"html/template"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	CloudwatchCredentialsTemplate = template.Must(template.New("cw credentials").Parse(cloudwatchCredentialsTemplateStr))
	//go:embed cloudwatch.credentials.tmpl
	cloudwatchCredentialsTemplateStr string
)

// ProfileCredentials contains fields for populating an AWS credentials file with profile info
type ProfileCredentials struct {
	Name                 string
	RoleARN              string
	WebIdentityTokenFile string
	AssumeRole           string
	ExternalID           string
	SessionName          string
}

func (e ProfileCredentials) Template() string {
	return `{{define "` + e.Name + `" -}}
{{range . -}}
{{if .AssumeRole -}}
# Source profile for initial web identity authentication
[output_{{ .Name }}_source]
web_identity_token_file={{ .WebIdentityTokenFile }}
role_arn={{ .RoleARN }}
role_session_name=output-{{ .Name }}-source

# Assume role profile for cross-account access
[output_{{ .Name }}]
source_profile=output_{{ .Name }}_source
role_arn={{ .AssumeRole }}
{{- if .ExternalID }}
external_id={{ .ExternalID }}
{{- end }}
{{- if .SessionName }}
role_session_name={{ .SessionName }}
{{- else }}
role_session_name=output-{{ .Name }}
{{- end }}
{{else -}}
# Direct authentication profile
[output_{{ .Name }}]
web_identity_token_file={{ .WebIdentityTokenFile }}
role_arn={{ .RoleARN }}
role_session_name=output-{{ .Name }}
{{end}}

{{- end}}
`
}

// IsCloudwatchRoleAuth checks if any output contains role authentication
// Used to determine if a configMap credentials file should be created
func IsCloudwatchRoleAuth(outputs []obs.OutputSpec) bool {
	for _, o := range outputs {
		if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.IAMRole != nil {
			return true
		}
	}
	return false
}

// ReconcileCredentialsFile reconciles a configmap with credential profiles for aws outputs
func ReconcileCredentialsFile(k8sClient client.Client, reader client.Reader, namespace, name, clfName string, outputs []obs.OutputSpec, secrets observability.Secrets, configMaps map[string]*corev1.ConfigMap, owner metav1.OwnerReference) (*corev1.ConfigMap, error) {
	log.V(5).Info("generating AWS ConfigMap")
	credString, err := GenerateCredentials(reader, clfName, outputs, secrets)

	if err != nil || credString == "" {
		return nil, err
	}

	credsConfMap := runtime.NewConfigMap(
		namespace,
		name,
		map[string]string{
			constants.AWSCredentialsKey: credString,
		})
	utils.AddOwnerRefToObject(credsConfMap, owner)

	return credsConfMap, reconcile.Configmap(k8sClient, k8sClient, credsConfMap)
}

// GenerateCredentials generates IAM profile credentials as a formatted string.
func GenerateCredentials(reader client.Reader, clfName string, outputs []obs.OutputSpec, secrets observability.Secrets) (string, error) {
	credentials := GenerateProfileCredentials(reader, clfName, outputs, secrets)
	if credentials == nil {
		return "", nil
	}

	// Execute Go template to generate credential profile(s)
	w := &strings.Builder{}
	err := CloudwatchCredentialsTemplate.Execute(w, credentials)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

// GenerateProfileCredentials generates a credential object from secrets and projected tokens
func GenerateProfileCredentials(reader client.Reader, clfName string, outputs []obs.OutputSpec, secrets observability.Secrets) (profileCredentials []ProfileCredentials) {
	for _, o := range outputs {
		if o.Cloudwatch != nil && o.Cloudwatch.Authentication.IAMRole != nil {
			roleAuth := o.Cloudwatch.Authentication.IAMRole
			if roleAuth != nil {
				assumeRole := secrets.AsString(&roleAuth.RoleARN)
				tokenPath := common.ServiceAccountBasePath(constants.TokenKey)

				if roleAuth.Token.From == obs.BearerTokenFromSecret {
					secret := roleAuth.Token.Secret
					tokenPath = common.SecretPath(secret.Name, secret.Key)
				}

				// Build credentials objects
				creds := ProfileCredentials{
					Name:                 o.Name,
					RoleARN:              assumeRole,
					WebIdentityTokenFile: tokenPath,
				}

				// Assume role cross-account auth
				if roleAuth.AssumeRole != nil {
					creds.AssumeRole = secrets.AsString(&roleAuth.AssumeRole.RoleARN)
					if roleAuth.AssumeRole.ExternalID != "" {
						creds.ExternalID = roleAuth.AssumeRole.ExternalID
					}
					// Use intelligent session name generation with cluster metadata
					creds.SessionName = generateSessionName(reader, clfName, o.Name)
				}
				profileCredentials = append(profileCredentials, creds)
			}
		}
	}
	return profileCredentials
}

// generateSessionName creates an intelligent session name using cluster metadata
// Format: <cluster_ID>-<CLF_Name>-<output_name>
func generateSessionName(reader client.Reader, clfName, outputName string) string {
	// Try to get cluster ID for meaningful session names
	_, clusterID, err := version.ClusterVersion(reader)
	if err != nil || clusterID == "" {
		// Fallback to basic session name if cluster metadata unavailable
		log.V(3).Info("Unable to retrieve cluster ID for session name, using fallback", "error", err)
		return "output-" + outputName
	}

	// Generate session name with cluster context for better auditing
	// Format: clusterid-clfname-output (truncated to 64 chars max)
	sessionName := clusterID[:min(8, len(clusterID))] + "-" + clfName + "-" + outputName
	if len(sessionName) > 64 {
		// Truncate but keep meaningful parts, prioritizing cluster ID and output name
		sessionName = clusterID[:min(8, len(clusterID))] + "-" + outputName
		if len(sessionName) > 64 {
			sessionName = sessionName[:64]
		}
	}
	return sessionName
}

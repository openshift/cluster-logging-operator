package cloudwatch

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"html/template"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	CloudwatchCredentialsTemplate = template.Must(template.New("cw credentials").Parse(cloudwatchCredentialsTemplateStr))
	//go:embed cloudwatch.credentials.tmpl
	cloudwatchCredentialsTemplateStr string
)

type CloudwatchWebIdentity struct {
	Name                 string
	RoleARN              string
	WebIdentityTokenFile string
	// Assume role configuration
	AssumeRoleARN string
	ExternalID    string
	SessionName   string
}

// generateSessionName creates an intelligent session name using cluster metadata
// Format: <cluster_ID>-<namespace>-<CLF_Name>-<output_name>
// Uses hash-based truncation to ensure uniqueness when names exceed AWS 64-character limit
func generateSessionName(clusterID, namespace, clfName, outputName string) string {
	var fullName string
	
	if clusterID == "" {
		// Use recommended fallback format: {namespace}-{clfName}-{outputName}
		log.V(3).Info("Cluster ID not available for session name, using namespace-clf-output format")
		fullName = namespace + "-" + clfName + "-" + outputName
	} else {
		// Use full format with truncated cluster ID
		clusterPrefix := clusterID
		if len(clusterID) > 8 {
			clusterPrefix = clusterID[:8]
		}
		fullName = clusterPrefix + "-" + namespace + "-" + clfName + "-" + outputName
	}
	
	// AWS session names have a 64-character limit
	if len(fullName) <= 64 {
		return fullName
	}
	
	// Hash-based truncation for uniqueness when exceeding limit
	hash := sha256.Sum256([]byte(fullName))
	hashSuffix := hex.EncodeToString(hash[:])[:8] // 8-character hash
	
	// Reserve space for hash suffix and separator
	maxPrefixLength := 64 - len(hashSuffix) - 1 // 64 total - 8 hash - 1 dash
	
	// Truncate at meaningful boundary (prefer keeping cluster ID and output name)
	truncatedPrefix := fullName
	if len(fullName) > maxPrefixLength {
		truncatedPrefix = fullName[:maxPrefixLength]
	}
	
	result := truncatedPrefix + "-" + hashSuffix
	log.V(3).Info("Session name truncated with hash", "original", fullName, "truncated", result)
	
	return result
}

// ReconcileAWSCredentialsConfigMap reconciles a configmap with credential profile(s) for Cloudwatch output(s).
func ReconcileAWSCredentialsConfigMap(k8sClient client.Client, reader client.Reader, namespace, name, clfName, clusterID string, outputs []obs.OutputSpec, secrets observability.Secrets, configMaps map[string]*corev1.ConfigMap, owner metav1.OwnerReference) (*corev1.ConfigMap, error) {
	log.V(3).Info("generating AWS ConfigMap")
	credString, err := GenerateCloudwatchCredentialProfiles(reader, namespace, clfName, clusterID, outputs, secrets)

	if err != nil || credString == "" {
		return nil, err
	}

	configMap := runtime.NewConfigMap(
		namespace,
		name,
		map[string]string{
			constants.AWSCredentialsKey: credString,
		})

	utils.AddOwnerRefToObject(configMap, owner)
	return configMap, reconcile.Configmap(k8sClient, k8sClient, configMap)
}

// GenerateCloudwatchCredentialProfiles generates AWS CLI profiles for a credentials file from spec'd cloudwatch role ARNs and returns the formatted content as a string.
func GenerateCloudwatchCredentialProfiles(reader client.Reader, namespace, clfName, clusterID string, outputs []obs.OutputSpec, secrets observability.Secrets) (string, error) {
	// Gather all cloudwatch output's role_arns/tokens
	webIds := GatherAWSWebIdentities(reader, namespace, clfName, clusterID, outputs, secrets)

	// No CW outputs
	if webIds == nil {
		return "", nil
	}

	// Execute Go template to generate credential profile(s)
	w := &strings.Builder{}
	err := CloudwatchCredentialsTemplate.Execute(w, webIds)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

// GatherAWSWebIdentities takes spec'd role arns and generates CloudwatchWebIdentity objects with a name and token path from secret or projected SA token
func GatherAWSWebIdentities(reader client.Reader, namespace, clfName, clusterID string, outputs []obs.OutputSpec, secrets observability.Secrets) (webIds []CloudwatchWebIdentity) {
	for _, o := range outputs {
		if o.Type == obs.OutputTypeCloudwatch && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.Type == obs.CloudwatchAuthTypeIAMRole {
			if roleARN := cloudwatch.ParseRoleArn(o.Cloudwatch.Authentication, secrets); roleARN != "" {
				tokenPath := common.ServiceAccountBasePath(constants.TokenKey)
				if o.Cloudwatch.Authentication.IAMRole.Token.From == obs.BearerTokenFromSecret {
					secret := o.Cloudwatch.Authentication.IAMRole.Token.Secret
					tokenPath = common.SecretPath(secret.Name, secret.Key)
				}

				webId := CloudwatchWebIdentity{
					Name:                 o.Name,
					RoleARN:              roleARN,
					WebIdentityTokenFile: tokenPath,
				}

				// Add assume role configuration if specified
				if o.Cloudwatch.Authentication.IAMRole != nil && o.Cloudwatch.Authentication.IAMRole.AssumeRole != nil {
					if assumeRoleARN := cloudwatch.ParseAssumeRoleArn(o.Cloudwatch.Authentication, secrets); assumeRoleARN != "" {
						webId.AssumeRoleARN = assumeRoleARN
					}
					if o.Cloudwatch.Authentication.IAMRole.AssumeRole.ExternalID != nil {
						webId.ExternalID = secrets.AsString(o.Cloudwatch.Authentication.IAMRole.AssumeRole.ExternalID)
					}
					// Use intelligent session name generation with cluster metadata
					webId.SessionName = generateSessionName(clusterID, namespace, clfName, o.Name)
				}

				webIds = append(webIds, webId)
			}
		}
	}
	return webIds
}

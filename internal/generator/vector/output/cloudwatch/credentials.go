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
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	ProfileTemplate = template.Must(template.New("creds").Parse(profileCreds))
	//go:embed cloudwatch.credentials.tmpl
	profileCreds string
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

// ReconcileCredsFile reconciles a configmap with credential profiles for aws outputs
func ReconcileCredsFile(k8sClient client.Client, reader client.Reader, namespace, name, clfName string, outputs []obs.OutputSpec, secrets observability.Secrets, configMaps map[string]*corev1.ConfigMap, owner metav1.OwnerReference) (*corev1.ConfigMap, error) {
	log.V(5).Info("generating AWS ConfigMap")
	credString, err := GenerateCredString(reader, clfName, outputs, secrets)
	if err != nil || credString == "" {
		return nil, err
	}
	profileConfigMap := runtime.NewConfigMap(
		namespace,
		name,
		map[string]string{
			constants.AWSCredentialsKey: credString,
		})
	utils.AddOwnerRefToObject(profileConfigMap, owner)
	return profileConfigMap, reconcile.Configmap(k8sClient, k8sClient, profileConfigMap)
}

// GenerateCredString generates IAM profile credentials as a formatted string
func GenerateCredString(reader client.Reader, clfName string, outputs []obs.OutputSpec, secrets observability.Secrets) (string, error) {
	credsData := GenerateProfileCreds(reader, clfName, outputs, secrets)
	if credsData == nil {
		return "", nil
	}
	// Generate cred profile string using go template
	w := &strings.Builder{}
	err := ProfileTemplate.Execute(w, credsData)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

// GenerateProfileCreds generates a credential object from secrets and projected tokens for each output
func GenerateProfileCreds(reader client.Reader, clfName string, outputs []obs.OutputSpec, secrets observability.Secrets) (profileCredentials []ProfileCredentials) {
	for _, o := range outputs {
		if isRoleAuth, awsAuth := OutputIsRoleAuth(o); isRoleAuth {
			// Build credentials object
			profile := ProfileCredentials{
				Name: o.Name,
				// we are using parse role here to accommodate ccoctl implementations of the secret
				RoleARN:              ParseRoleArn(awsAuth, secrets),
				WebIdentityTokenFile: common.ServiceAccountBasePath(constants.TokenKey),
			}
			// Not using projected token
			if awsAuth.IAMRole.Token.From == obs.BearerTokenFromSecret {
				secret := awsAuth.IAMRole.Token.Secret
				profile.WebIdentityTokenFile = common.SecretPath(secret.Name, secret.Key)
			}
			// Add any cross-account assumeRole
			if isAssumeRole, assumeRoleSpec := OutputIsAssumeRole(o); isAssumeRole {
				profile.AssumeRole = secrets.AsString(&assumeRoleSpec.RoleARN)
				if hasExtID, extID := AssumeRoleHasExternalId(assumeRoleSpec); hasExtID {
					profile.ExternalID = extID
				}
				// Use intelligent session name generation
				profile.SessionName = generateSessionName(reader, clfName, o.Name)
			}
			profileCredentials = append(profileCredentials, profile)
		}
	}
	return profileCredentials
}

// generateSessionName creates an intelligent session name using cluster metadata
func generateSessionName(reader client.Reader, clfName, outputName string) string {
	// Get cluster ID for meaningful session names
	_, clusterID, err := version.ClusterVersion(reader)
	if err != nil || clusterID == "" {
		// Fallback to basic session name if cluster metadata unavailable
		log.V(3).Info("Unable to retrieve cluster ID for session name, using fallback", "error", err)
		return "output-" + outputName
	}
	// Format: clusterid-clfname-output (truncated to 64 chars max)
	sessionName := clusterID[:min(8, len(clusterID))] + "-" + clfName + "-" + outputName
	if len(sessionName) > 64 {
		// Truncate but prioritizing cluster ID and output name
		sessionName = clusterID[:min(8, len(clusterID))] + "-" + outputName
		if len(sessionName) > 64 {
			sessionName = sessionName[:64]
		}
	}
	return sessionName
}

// RequiresProfilesConfigMap determine if a credentials configMap should be created for AWS
func RequiresProfilesConfigMap(outputs []obs.OutputSpec) bool {
	for _, o := range outputs {
		if found, _ := OutputIsRoleAuth(o); found {
			return true
		}
	}
	return false
}

// OutputIsRoleAuth identifies if `output.Cloudwatch.Authentication.IAMRole` exists and returns ref if so
func OutputIsRoleAuth(o obs.OutputSpec) (bool, *obs.CloudwatchAuthentication) {
	if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.IAMRole != nil {
		return true, o.Cloudwatch.Authentication
	}
	return false, nil
}

// OutputIsAssumeRole identifies if 'output.Cloudwatch.Authentication.AssumeRole` exists and returns ref if so
func OutputIsAssumeRole(o obs.OutputSpec) (bool, *obs.CloudwatchAssumeRole) {
	if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.AssumeRole != nil {
		return true, o.Cloudwatch.Authentication.AssumeRole
	}
	return false, nil
}

// AssumeRoleHasExternalId identifies if externalID exists and returns the string
func AssumeRoleHasExternalId(assumeRole *obs.CloudwatchAssumeRole) (bool, string) {
	if assumeRole.ExternalID != "" {
		return true, assumeRole.ExternalID
	}
	return false, ""
}

// ParseRoleArn search for valid AWS arn, return emtpy for no match
func ParseRoleArn(authSpec *obs.CloudwatchAuthentication, secrets observability.Secrets) string {
	var roleString string
	if authSpec.IAMRole != nil {
		roleString = secrets.AsString(&authSpec.IAMRole.RoleARN)
	}
	return findSubstring(roleString)
}

// ParseAssumeRoleArn search for valid AWS assumeRole arn, return empty for no match
func ParseAssumeRoleArn(assumeRoleSpec *obs.CloudwatchAssumeRole, secrets observability.Secrets) string {
	var roleString string
	if assumeRoleSpec != nil {
		roleString = secrets.AsString(&assumeRoleSpec.RoleARN)
	}
	return findSubstring(roleString)
}

// findSubstring matches regex on a valid AWS role arn and returns empty for no match
func findSubstring(roleString string) string {
	if roleString != "" {
		reg := regexp.MustCompile(`(arn:aws(.*)?:(iam|sts)::\d{12}:role\/\S+)\s?`)
		roleArn := reg.FindStringSubmatch(roleString)
		if roleArn != nil {
			return roleArn[1] // the capturing group is index 1
		}
	}
	return ""
}

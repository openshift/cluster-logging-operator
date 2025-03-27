package cloudwatch

import (
	_ "embed"
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
}

// ReconcileAWSCredentialsConfigMap reconciles a configmap with credential profile(s) for Cloudwatch output(s).
func ReconcileAWSCredentialsConfigMap(k8sClient client.Client, reader client.Reader, namespace, name string, outputs []obs.OutputSpec, secrets observability.Secrets, configMaps map[string]*corev1.ConfigMap, owner metav1.OwnerReference) (*corev1.ConfigMap, error) {
	log.V(3).Info("generating AWS ConfigMap")
	credString, err := GenerateCloudwatchCredentialProfiles(outputs, secrets)

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
func GenerateCloudwatchCredentialProfiles(outputs []obs.OutputSpec, secrets observability.Secrets) (string, error) {
	// Gather all cloudwatch output's role_arns/tokens
	webIds := GatherAWSWebIdentities(outputs, secrets)

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
func GatherAWSWebIdentities(outputs []obs.OutputSpec, secrets observability.Secrets) (webIds []CloudwatchWebIdentity) {
	for _, o := range outputs {
		if o.Type == obs.OutputTypeCloudwatch && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.Type == obs.CloudwatchAuthTypeIAMRole {
			if roleARN := cloudwatch.ParseRoleArn(o.Cloudwatch.Authentication, secrets); roleARN != "" {
				tokenPath := common.ServiceAccountBasePath(constants.TokenKey)
				if o.Cloudwatch.Authentication.IAMRole.Token.From == obs.BearerTokenFromSecret {
					secret := o.Cloudwatch.Authentication.IAMRole.Token.Secret
					tokenPath = common.SecretPath(secret.Name, secret.Key)
				}
				webIds = append(webIds, CloudwatchWebIdentity{
					Name:                 o.Name,
					RoleARN:              roleARN,
					WebIdentityTokenFile: tokenPath,
				})
			}
		}
	}
	return webIds
}

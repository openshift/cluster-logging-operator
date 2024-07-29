package api

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultEsName        = "default-elasticsearch"
	DefaultLokistackName = "default-lokistack"
	DefaultName          = "default-"
)

func convertOutputs(loggingClfSpec *logging.ClusterLogForwarderSpec, secrets map[string]*corev1.Secret) []obs.OutputSpec {
	obsOutputs := []obs.OutputSpec{}
	for _, output := range loggingClfSpec.Outputs {

		obsOut := &obs.OutputSpec{}
		obsOut.Name = output.Name
		obsOut.Type = obs.OutputType(output.Type)

		switch obsOut.Type {
		case obs.OutputTypeAzureMonitor:
			obsOut.AzureMonitor = mapAzureMonitor(output, secrets[output.Name])
		case obs.OutputTypeCloudwatch:
			obsOut.Cloudwatch = mapCloudwatch(output, secrets[output.Name])
		case obs.OutputTypeElasticsearch:
			obsOut.Elasticsearch = mapElasticsearch(output, secrets[output.Name])
		case obs.OutputTypeGoogleCloudLogging:
			obsOut.GoogleCloudLogging = mapGoogleCloudLogging(output, secrets[output.Name])
		case obs.OutputTypeHTTP:
			obsOut.HTTP = mapHTTP(output, secrets[output.Name])
		case obs.OutputTypeKafka:
			obsOut.Kafka = mapKafka(output, secrets[output.Name])
		case obs.OutputTypeLoki:
			obsOut.Loki = mapLoki(output, secrets[output.Name])
		case obs.OutputTypeSplunk:
			obsOut.Splunk = mapSplunk(output, secrets[output.Name])
		case obs.OutputTypeSyslog:
			obsOut.Syslog = mapSyslog(output)
		}
		// Limits
		if output.Limit != nil {
			obsOut.Limit = &obs.LimitSpec{
				MaxRecordsPerSecond: output.Limit.MaxRecordsPerSecond,
			}
		}

		// TLS Settings
		if output.TLS != nil {
			obsOut.TLS = mapOutputTls(output.TLS, secrets[output.Name])
		}

		// Add output to obs clf
		obsOutputs = append(obsOutputs, *obsOut)
	}
	// Set observability CLF outputs to converted outputs
	return obsOutputs
}

// ReferencesFluentDForward determines if FluentDForward is a defined output
func ReferencesFluentDForward(loggingClfSpec *logging.ClusterLogForwarderSpec) bool {
	for _, output := range loggingClfSpec.Outputs {
		if output.Type == logging.OutputTypeFluentdForward {
			return true
		}
	}
	return false
}

func generateDefaultOutput(logStoreSpec *logging.LogStoreSpec) *obs.OutputSpec {
	var output *obs.OutputSpec
	var outputName string

	switch logStoreSpec.Type {
	case logging.LogStoreTypeElasticsearch:
		outputName = DefaultEsName
		output = &obs.OutputSpec{
			Name: outputName,
			Type: obs.OutputTypeElasticsearch,
			Elasticsearch: &obs.Elasticsearch{
				URLSpec: obs.URLSpec{
					URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
				},
				Version: 6,
				Index:   `{.log_type||"none"}`,
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: constants.ElasticsearchName,
					},
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: constants.ElasticsearchName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: constants.ElasticsearchName,
					},
				},
			},
		}
	case logging.LogStoreTypeLokiStack:
		outputName = DefaultLokistackName
		output = &obs.OutputSpec{
			Name: outputName,
			Type: obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      logStoreSpec.LokiStack.Name,
					Namespace: constants.OpenshiftNS,
				},
				Authentication: &obs.LokiStackAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Name: constants.LogCollectorToken,
							Key:  constants.BearerTokenFileKey,
						},
					},
				},
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        "service-ca.crt",
						SecretName: constants.LogCollectorToken,
					},
				},
			},
		}
	}
	return output
}

func mapOutputTls(loggingTls *logging.OutputTLSSpec, outputSecret *corev1.Secret) *obs.OutputTLSSpec {
	obsTls := &obs.OutputTLSSpec{
		InsecureSkipVerify: loggingTls.InsecureSkipVerify,
		TLSSecurityProfile: loggingTls.TLSSecurityProfile,
	}

	if security.HasTLSCertAndKey(outputSecret) {
		obsTls.Certificate = &obs.ValueReference{
			Key:        constants.ClientCertKey,
			SecretName: outputSecret.Name,
		}
		obsTls.Key = &obs.SecretReference{
			Key:        constants.ClientPrivateKey,
			SecretName: outputSecret.Name,
		}
	}
	if security.HasCABundle(outputSecret) {
		obsTls.CA = &obs.ValueReference{
			Key:        constants.TrustedCABundleKey,
			SecretName: outputSecret.Name,
		}
	}
	if security.HasPassphrase(outputSecret) {
		obsTls.KeyPassphrase = &obs.SecretReference{
			Key:        constants.Passphrase,
			SecretName: outputSecret.Name,
		}
	}

	return obsTls
}
func mapBaseOutputTuning(outTuneSpec logging.OutputTuningSpec) *obs.BaseOutputTuningSpec {
	obsBaseTuningSpec := &obs.BaseOutputTuningSpec{
		MaxWrite:         outTuneSpec.MaxWrite,
		MinRetryDuration: outTuneSpec.MinRetryDuration,
		MaxRetryDuration: outTuneSpec.MaxRetryDuration,
	}

	switch outTuneSpec.Delivery {
	case logging.OutputDeliveryModeAtLeastOnce:
		obsBaseTuningSpec.Delivery = obs.DeliveryModeAtLeastOnce
	case logging.OutputDeliveryModeAtMostOnce:
		obsBaseTuningSpec.Delivery = obs.DeliveryModeAtMostOnce
	}

	return obsBaseTuningSpec
}

func mapHTTPAuth(secret *corev1.Secret) *obs.HTTPAuthentication {
	httpAuth := obs.HTTPAuthentication{}
	if security.HasUsernamePassword(secret) {
		httpAuth.Username = &obs.SecretReference{
			Key:        constants.ClientUsername,
			SecretName: secret.Name,
		}
		httpAuth.Password = &obs.SecretReference{
			Key:        constants.ClientPassword,
			SecretName: secret.Name,
		}
	}
	if security.HasBearerTokenFileKey(secret) {
		httpAuth.Token = &obs.BearerToken{
			From: obs.BearerTokenFromSecret,
			Secret: &obs.BearerTokenSecretKey{
				Name: secret.Name,
				Key:  constants.BearerTokenFileKey,
			},
		}
	}
	return &httpAuth
}

func mapAzureMonitor(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.AzureMonitor {
	obsAzMon := &obs.AzureMonitor{}

	// Authentication
	if secret != nil {
		obsAzMon.Authentication = &obs.AzureMonitorAuthentication{}
		if security.HasSharedKey(secret) {
			obsAzMon.Authentication.SharedKey = &obs.SecretReference{
				Key:        constants.SharedKey,
				SecretName: secret.Name,
			}
		}
	}

	// Tuning Specs
	if loggingOutSpec.Tuning != nil {
		obsAzMon.Tuning = mapBaseOutputTuning(*loggingOutSpec.Tuning)
	}

	loggingAzMon := loggingOutSpec.AzureMonitor
	if loggingAzMon == nil {
		return obsAzMon
	}

	obsAzMon.CustomerId = loggingAzMon.CustomerId
	obsAzMon.LogType = loggingAzMon.LogType
	obsAzMon.AzureResourceId = loggingAzMon.AzureResourceId
	obsAzMon.Host = loggingAzMon.Host

	return obsAzMon
}

func mapCloudwatch(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Cloudwatch {
	obsCw := &obs.Cloudwatch{}

	if loggingOutSpec.URL != "" {
		obsCw.URL = &loggingOutSpec.URL
	}

	// Map secret to authentication
	if secret != nil {
		obsCw.Authentication = &obs.CloudwatchAuthentication{}
		if security.HasAwsAccessKeyId(secret) && security.HasAwsSecretAccessKey(secret) {
			obsCw.Authentication.Type = obs.CloudwatchAuthTypeAccessKey
			obsCw.Authentication.AWSAccessKey = &obs.CloudwatchAWSAccessKey{
				KeyID: &obs.SecretReference{
					Key:        constants.AWSAccessKeyID,
					SecretName: secret.Name,
				},
				KeySecret: &obs.SecretReference{
					Key:        constants.AWSSecretAccessKey,
					SecretName: secret.Name,
				},
			}
		}

		if security.HasAwsRoleArnKey(secret) || security.HasAwsCredentialsKey(secret) {
			obsCw.Authentication.Type = obs.CloudwatchAuthTypeIAMRole

			// Determine if `role_arn` or `credentials` key is specified
			roleArnKey := constants.AWSWebIdentityRoleKey
			if security.HasAwsCredentialsKey(secret) {
				roleArnKey = constants.AWSCredentialsKey
			}

			obsCw.Authentication.IAMRole = &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretReference{
					Key:        roleArnKey,
					SecretName: secret.Name,
				},
			}
			if security.HasBearerTokenFileKey(secret) {
				obsCw.Authentication.IAMRole.Token = &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{
						Name: secret.Name,
						Key:  constants.BearerTokenFileKey,
					},
				}
			} else {
				obsCw.Authentication.IAMRole.Token = &obs.BearerToken{
					From: obs.BearerTokenFromServiceAccount,
				}
			}
		}
	}

	if loggingOutSpec.Tuning != nil {
		obsCw.Tuning = &obs.CloudwatchTuningSpec{
			Compression:          loggingOutSpec.Tuning.Compression,
			BaseOutputTuningSpec: *mapBaseOutputTuning(*loggingOutSpec.Tuning),
		}
	}

	loggingCw := loggingOutSpec.Cloudwatch
	if loggingCw == nil {
		return obsCw
	}

	obsCw.Region = loggingCw.Region

	// Group name
	groupBy := ""
	switch loggingCw.GroupBy {
	case logging.LogGroupByLogType:
		groupBy = ".log_type"
	case logging.LogGroupByNamespaceName:
		groupBy = ".kubernetes.namespace_name"
	case logging.LogGroupByNamespaceUUID:
		groupBy = ".kubernetes.namespace_uid"
	}
	groupPrefix := ""

	if loggingCw.GroupPrefix != nil {
		groupPrefix = *loggingCw.GroupPrefix
	} else {
		groupPrefix = `{.openshift.cluster_id||"none"}`
	}

	obsCw.GroupName = fmt.Sprintf(`%s.{%s||"none"}`, groupPrefix, groupBy)

	return obsCw
}

func mapElasticsearch(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Elasticsearch {
	obsEs := &obs.Elasticsearch{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
		Version: 8,
		Index:   `{.log_type||"none"}`,
	}

	if secret != nil {
		obsEs.Authentication = mapHTTPAuth(secret)
	}

	if loggingOutSpec.Tuning != nil {
		obsEs.Tuning = &obs.ElasticsearchTuningSpec{
			BaseOutputTuningSpec: *mapBaseOutputTuning(*loggingOutSpec.Tuning),
			Compression:          loggingOutSpec.Tuning.Compression,
		}
	}

	loggingES := loggingOutSpec.Elasticsearch
	if loggingES == nil {
		return obsEs
	}

	obsEs.Version = loggingES.Version

	if loggingES.StructuredTypeKey != "" && loggingES.StructuredTypeName != "" {
		// Ensure StructuredTypeKey begins with `.`
		structuredTypeKey := loggingES.StructuredTypeKey
		if !strings.HasPrefix(structuredTypeKey, ".") {
			structuredTypeKey = "." + structuredTypeKey
		}
		obsEs.Index = fmt.Sprintf("{%s||%q}", structuredTypeKey, loggingES.StructuredTypeName)
	}

	return obsEs
}

func mapGoogleCloudLogging(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.GoogleCloudLogging {
	obsGcp := &obs.GoogleCloudLogging{}

	if secret != nil {
		obsGcp.Authentication = &obs.GoogleCloudLoggingAuthentication{}
		if security.HasGoogleApplicationCredentialsKey(secret) {
			obsGcp.Authentication.Credentials = &obs.SecretReference{
				Key:        gcl.GoogleApplicationCredentialsKey,
				SecretName: secret.Name,
			}
		}
	}
	if loggingOutSpec.Tuning != nil {
		obsGcp.Tuning = &obs.GoogleCloudLoggingTuningSpec{
			BaseOutputTuningSpec: *mapBaseOutputTuning(*loggingOutSpec.Tuning),
		}
	}

	loggingGcp := loggingOutSpec.GoogleCloudLogging
	if loggingGcp == nil {
		return obsGcp
	}
	if loggingGcp.BillingAccountID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeBillingAccount,
			Value: loggingGcp.BillingAccountID,
		}
	} else if loggingGcp.OrganizationID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeOrganization,
			Value: loggingGcp.OrganizationID,
		}
	} else if loggingGcp.FolderID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeFolder,
			Value: loggingGcp.FolderID,
		}
	} else if loggingGcp.ProjectID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeProject,
			Value: loggingGcp.ProjectID,
		}
	}

	obsGcp.LogID = loggingGcp.LogID

	return obsGcp
}

func mapHTTP(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.HTTP {
	obsHTTP := &obs.HTTP{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
	}

	if secret != nil {
		obsHTTP.Authentication = mapHTTPAuth(secret)
	}

	if loggingOutSpec.Tuning != nil {
		obsHTTP.Tuning = &obs.HTTPTuningSpec{
			BaseOutputTuningSpec: *mapBaseOutputTuning(*loggingOutSpec.Tuning),
			Compression:          loggingOutSpec.Tuning.Compression,
		}
	}

	loggingHTTP := loggingOutSpec.Http
	if loggingHTTP == nil {
		return obsHTTP
	}

	obsHTTP.Headers = loggingHTTP.Headers
	obsHTTP.Timeout = loggingHTTP.Timeout
	obsHTTP.Method = loggingHTTP.Method

	return obsHTTP
}

func mapKafka(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Kafka {
	obsKafka := &obs.Kafka{}

	if loggingOutSpec.URL != "" {
		obsKafka.URL = &loggingOutSpec.URL
	}

	if secret != nil {
		obsKafka.Authentication = &obs.KafkaAuthentication{
			SASL: &obs.SASLAuthentication{Mechanism: "PLAIN"},
		}
		if security.HasUsernamePassword(secret) {
			obsKafka.Authentication.SASL.Username = &obs.SecretReference{
				Key:        constants.ClientUsername,
				SecretName: secret.Name,
			}
			obsKafka.Authentication.SASL.Password = &obs.SecretReference{
				Key:        constants.ClientPassword,
				SecretName: secret.Name,
			}
		}
		if security.HasSASLMechanism(secret) {
			if m := security.GetFromSecret(secret, constants.SASLMechanisms); m != "" {
				obsKafka.Authentication.SASL.Mechanism = m
			}
		}
	}

	if loggingOutSpec.Tuning != nil {
		obsKafka.Tuning = &obs.KafkaTuningSpec{
			MaxWrite:    loggingOutSpec.Tuning.MaxWrite,
			Compression: loggingOutSpec.Tuning.Compression,
		}

		switch loggingOutSpec.Tuning.Delivery {
		case logging.OutputDeliveryModeAtLeastOnce:
			obsKafka.Tuning.Delivery = obs.DeliveryModeAtLeastOnce
		case logging.OutputDeliveryModeAtMostOnce:
			obsKafka.Tuning.Delivery = obs.DeliveryModeAtMostOnce
		}
	}

	loggingKafka := loggingOutSpec.Kafka
	if loggingKafka == nil {
		return obsKafka
	}

	obsKafka.Topic = loggingKafka.Topic
	for _, b := range loggingKafka.Brokers {
		obsKafka.Brokers = append(obsKafka.Brokers, obs.URL(b))
	}

	return obsKafka
}

func mapLoki(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Loki {
	obsLoki := &obs.Loki{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
	}

	if secret != nil {
		obsLoki.Authentication = mapHTTPAuth(secret)
	}

	if loggingOutSpec.Tuning != nil {
		obsLoki.Tuning = &obs.LokiTuningSpec{
			BaseOutputTuningSpec: *mapBaseOutputTuning(*loggingOutSpec.Tuning),
			Compression:          loggingOutSpec.Tuning.Compression,
		}
	}

	loggingLoki := loggingOutSpec.Loki
	if loggingLoki == nil {
		return obsLoki
	}

	if loggingLoki.TenantKey != "" {
		tenantKey := loggingLoki.TenantKey
		if !strings.HasPrefix(tenantKey, ".") {
			tenantKey = "." + tenantKey
		}
		obsLoki.TenantKey = fmt.Sprintf(`{%s||"none"}`, tenantKey)
	}
	obsLoki.LabelKeys = loggingLoki.LabelKeys

	return obsLoki
}

func mapSplunk(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Splunk {
	obsSplunk := &obs.Splunk{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
	}

	// Set auth
	if secret != nil {
		obsSplunk.Authentication = &obs.SplunkAuthentication{}
		if security.HasSplunkHecToken(secret) {
			obsSplunk.Authentication.Token = &obs.SecretReference{
				Key:        constants.SplunkHECTokenKey,
				SecretName: secret.Name,
			}
		}
	}

	// Set tuning
	if loggingOutSpec.Tuning != nil {
		obsSplunk.Tuning = &obs.SplunkTuningSpec{
			BaseOutputTuningSpec: *mapBaseOutputTuning(*loggingOutSpec.Tuning),
		}
	}

	loggingSplunk := loggingOutSpec.Splunk

	if loggingSplunk == nil {
		return obsSplunk
	}

	// Set index if specified
	var splunkIndex string
	if loggingSplunk.IndexKey != "" {
		indexKey := loggingSplunk.IndexKey
		if !strings.HasPrefix(indexKey, ".") {
			indexKey = "." + indexKey
		}
		splunkIndex = fmt.Sprintf(`{%s||""}`, indexKey)
	} else if loggingSplunk.IndexName != "" {
		splunkIndex = loggingSplunk.IndexName
	}
	obsSplunk.Index = splunkIndex

	return obsSplunk
}

func mapSyslog(loggingOutSpec logging.OutputSpec) *obs.Syslog {
	obsSyslog := &obs.Syslog{
		URL: loggingOutSpec.URL,
	}

	loggingSyslog := loggingOutSpec.Syslog
	if loggingSyslog == nil {
		return obsSyslog
	}

	obsSyslog.RFC = obs.SyslogRFCType(loggingSyslog.RFC)
	obsSyslog.Facility = loggingSyslog.Facility
	obsSyslog.Severity = loggingSyslog.Severity

	if loggingSyslog.AddLogSource {
		obsSyslog.Enrichment = obs.EnrichmentTypeKubernetesMinimal
	}

	obsSyslog.AppName = loggingSyslog.AppName
	if strings.HasPrefix(loggingSyslog.AppName, "$.message") {
		obsSyslog.AppName = fmt.Sprintf(`{%s||"none"}`, strings.Split(loggingSyslog.AppName, "$.message")[1])
	}

	obsSyslog.ProcID = loggingSyslog.ProcID
	if strings.HasPrefix(loggingSyslog.ProcID, "$.message") {
		obsSyslog.AppName = fmt.Sprintf(`{%s||"none"}`, strings.Split(loggingSyslog.ProcID, "$.message")[1])
	}

	obsSyslog.MsgID = loggingSyslog.MsgID
	if strings.HasPrefix(loggingSyslog.MsgID, "$.message") {
		obsSyslog.AppName = fmt.Sprintf(`{%s||"none"}`, strings.Split(loggingSyslog.MsgID, "$.message")[1])
	}

	if loggingSyslog.PayloadKey != "" {
		obsSyslog.PayloadKey = fmt.Sprintf("{.%s}", loggingSyslog.PayloadKey)
	}

	return obsSyslog
}

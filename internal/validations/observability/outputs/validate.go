package outputs

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
)

func Validate(context internalcontext.ForwarderContext) {
	pipelines := internalobs.Pipelines(context.Forwarder.Spec.Pipelines)
	for _, out := range context.Forwarder.Spec.Outputs {
		messages := []string{}
		configs := internalobs.SecretReferencesAsValueReferences(out)
		if out.TLS != nil {
			messages = append(messages, validateURLAccordingToTLS(out)...)
			configs = append(configs, internalobs.ValueReferences(out.TLS.TLSSpec)...)
		}
		messages = append(messages, common.ValidateValueReference(configs, context.Secrets, context.ConfigMaps)...)
		messages = append(messages, validateOutputIsReferencedByPipelines(out, pipelines)...)
		messages = append(messages, validateOutputTemplates(out)...)
		// Validate by output type
		switch out.Type {
		case obs.OutputTypeCloudwatch, obs.OutputTypeS3:
			messages = append(messages, ValidateAwsAuth(out, context)...)
			if out.Type == obs.OutputTypeCloudwatch {
				messages = append(messages, validateCloudwatchMaxWrite(out)...)
			}
		case obs.OutputTypeGoogleCloudLogging:
			messages = append(messages, ValidateGCLAuth(out, context)...)
		case obs.OutputTypeHTTP:
			messages = append(messages, validateHttpContentTypeHeaders(out)...)
		case obs.OutputTypeElasticsearch:
			messages = append(messages, validateElasticsearchHeaders(out)...)
		case obs.OutputTypeAzureLogsIngestion:
			messages = append(messages, validateAzureLogsIngestionMaxWrite(out)...)
		}
		// Set condition
		if len(messages) > 0 {
			internalobs.SetCondition(&context.Forwarder.Status.OutputConditions,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, out.Name, false, obs.ReasonValidationFailure, strings.Join(messages, ",")))
		} else {
			internalobs.SetCondition(&context.Forwarder.Status.OutputConditions,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, out.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("output %q is valid", out.Name)))
		}
	}
}

// validateOutputTemplates rejects user-supplied template fields that would allow TOML literal
// multiline injection. These fields are rendered into a VRL remap whose source is serialized as
// a single literal multiline TOML string, so an embedded terminator would break out and inject
// arbitrary TOML into the collector configuration. This is the admission-time counterpart to the
// structural guard in internal/utils/toml. See LOG-9752.
func validateOutputTemplates(output obs.OutputSpec) (results []string) {
	add := func(fieldName, value string) {
		if msg := common.ValidateTOMLLiteralSafe(fieldName, value); msg != "" {
			results = append(results, msg)
		}
	}
	if output.Cloudwatch != nil {
		add("cloudwatch.groupName", output.Cloudwatch.GroupName)
	}
	if output.Elasticsearch != nil {
		add("elasticsearch.index", output.Elasticsearch.Index)
	}
	if output.GoogleCloudLogging != nil {
		add("googleCloudLogging.logId", output.GoogleCloudLogging.LogId)
	}
	if output.Kafka != nil {
		add("kafka.topic", output.Kafka.Topic)
	}
	if output.Loki != nil {
		add("loki.tenantKey", output.Loki.TenantKey)
	}
	if output.S3 != nil {
		add("s3.keyPrefix", output.S3.KeyPrefix)
	}
	if output.Splunk != nil {
		add("splunk.index", output.Splunk.Index)
		add("splunk.source", output.Splunk.Source)
		add("splunk.sourceType", output.Splunk.SourceType)
		add("splunk.payloadKey", string(output.Splunk.PayloadKey))
	}
	if output.Syslog != nil {
		add("syslog.facility", output.Syslog.Facility)
		add("syslog.severity", output.Syslog.Severity)
		add("syslog.appName", output.Syslog.AppName)
		add("syslog.procId", output.Syslog.ProcId)
		add("syslog.msgId", output.Syslog.MsgId)
		add("syslog.payloadKey", output.Syslog.PayloadKey)
	}
	return results
}

func validateOutputIsReferencedByPipelines(output obs.OutputSpec, pipelines internalobs.Pipelines) (results []string) {
	if !pipelines.ReferenceOutput(output) {
		return append(results, "not referenced by any pipeline")
	}
	return results
}

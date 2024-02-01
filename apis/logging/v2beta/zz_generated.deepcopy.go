//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v2beta

import (
	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = new(InclusionSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = new(InclusionSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Application.
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Audit) DeepCopyInto(out *Audit) {
	*out = *in
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Audit.
func (in *Audit) DeepCopy() *Audit {
	if in == nil {
		return nil
	}
	out := new(Audit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cloudwatch) DeepCopyInto(out *Cloudwatch) {
	*out = *in
	if in.GroupPrefix != nil {
		in, out := &in.GroupPrefix, &out.GroupPrefix
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cloudwatch.
func (in *Cloudwatch) DeepCopy() *Cloudwatch {
	if in == nil {
		return nil
	}
	out := new(Cloudwatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterLogForwarder) DeepCopyInto(out *ClusterLogForwarder) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterLogForwarder.
func (in *ClusterLogForwarder) DeepCopy() *ClusterLogForwarder {
	if in == nil {
		return nil
	}
	out := new(ClusterLogForwarder)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterLogForwarder) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterLogForwarderList) DeepCopyInto(out *ClusterLogForwarderList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterLogForwarder, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterLogForwarderList.
func (in *ClusterLogForwarderList) DeepCopy() *ClusterLogForwarderList {
	if in == nil {
		return nil
	}
	out := new(ClusterLogForwarderList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterLogForwarderList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterLogForwarderSpec) DeepCopyInto(out *ClusterLogForwarderSpec) {
	*out = *in
	if in.Inputs != nil {
		in, out := &in.Inputs, &out.Inputs
		*out = make([]InputSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Outputs != nil {
		in, out := &in.Outputs, &out.Outputs
		*out = make([]OutputSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = make([]FilterSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Pipelines != nil {
		in, out := &in.Pipelines, &out.Pipelines
		*out = make([]PipelineSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterLogForwarderSpec.
func (in *ClusterLogForwarderSpec) DeepCopy() *ClusterLogForwarderSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterLogForwarderSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterLogForwarderStatus) DeepCopyInto(out *ClusterLogForwarderStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Inputs != nil {
		in, out := &in.Inputs, &out.Inputs
		*out = make(map[string]v1.Condition, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Outputs != nil {
		in, out := &in.Outputs, &out.Outputs
		*out = make(map[string]v1.Condition, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = make(map[string]v1.Condition, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Pipelines != nil {
		in, out := &in.Pipelines, &out.Pipelines
		*out = make(map[string]v1.Condition, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterLogForwarderStatus.
func (in *ClusterLogForwarderStatus) DeepCopy() *ClusterLogForwarderStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterLogForwarderStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CollectorSpec) DeepCopyInto(out *CollectorSpec) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CollectorSpec.
func (in *CollectorSpec) DeepCopy() *CollectorSpec {
	if in == nil {
		return nil
	}
	out := new(CollectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Elasticsearch) DeepCopyInto(out *Elasticsearch) {
	*out = *in
	out.ElasticsearchStructuredSpec = in.ElasticsearchStructuredSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Elasticsearch.
func (in *Elasticsearch) DeepCopy() *Elasticsearch {
	if in == nil {
		return nil
	}
	out := new(Elasticsearch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchStructuredSpec) DeepCopyInto(out *ElasticsearchStructuredSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchStructuredSpec.
func (in *ElasticsearchStructuredSpec) DeepCopy() *ElasticsearchStructuredSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchStructuredSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FilterSpec) DeepCopyInto(out *FilterSpec) {
	*out = *in
	in.FilterTypeSpec.DeepCopyInto(&out.FilterTypeSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FilterSpec.
func (in *FilterSpec) DeepCopy() *FilterSpec {
	if in == nil {
		return nil
	}
	out := new(FilterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FilterTypeSpec) DeepCopyInto(out *FilterTypeSpec) {
	*out = *in
	if in.KubeAPIAudit != nil {
		in, out := &in.KubeAPIAudit, &out.KubeAPIAudit
		*out = new(KubeAPIAudit)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FilterTypeSpec.
func (in *FilterTypeSpec) DeepCopy() *FilterTypeSpec {
	if in == nil {
		return nil
	}
	out := new(FilterTypeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FluentdForward) DeepCopyInto(out *FluentdForward) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FluentdForward.
func (in *FluentdForward) DeepCopy() *FluentdForward {
	if in == nil {
		return nil
	}
	out := new(FluentdForward)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GoogleCloudLogging) DeepCopyInto(out *GoogleCloudLogging) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GoogleCloudLogging.
func (in *GoogleCloudLogging) DeepCopy() *GoogleCloudLogging {
	if in == nil {
		return nil
	}
	out := new(GoogleCloudLogging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPReceiver) DeepCopyInto(out *HTTPReceiver) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPReceiver.
func (in *HTTPReceiver) DeepCopy() *HTTPReceiver {
	if in == nil {
		return nil
	}
	out := new(HTTPReceiver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Http) DeepCopyInto(out *Http) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Http.
func (in *Http) DeepCopy() *Http {
	if in == nil {
		return nil
	}
	out := new(Http)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InclusionSpec) DeepCopyInto(out *InclusionSpec) {
	*out = *in
	if in.Include != nil {
		in, out := &in.Include, &out.Include
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Exclude != nil {
		in, out := &in.Exclude, &out.Exclude
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InclusionSpec.
func (in *InclusionSpec) DeepCopy() *InclusionSpec {
	if in == nil {
		return nil
	}
	out := new(InclusionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Infrastructure) DeepCopyInto(out *Infrastructure) {
	*out = *in
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Infrastructure.
func (in *Infrastructure) DeepCopy() *Infrastructure {
	if in == nil {
		return nil
	}
	out := new(Infrastructure)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputSpec) DeepCopyInto(out *InputSpec) {
	*out = *in
	if in.Application != nil {
		in, out := &in.Application, &out.Application
		*out = new(Application)
		(*in).DeepCopyInto(*out)
	}
	if in.Infrastructure != nil {
		in, out := &in.Infrastructure, &out.Infrastructure
		*out = new(Infrastructure)
		(*in).DeepCopyInto(*out)
	}
	if in.Audit != nil {
		in, out := &in.Audit, &out.Audit
		*out = new(Audit)
		(*in).DeepCopyInto(*out)
	}
	if in.Receiver != nil {
		in, out := &in.Receiver, &out.Receiver
		*out = new(ReceiverSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Tuning != nil {
		in, out := &in.Tuning, &out.Tuning
		*out = new(InputTuningSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputSpec.
func (in *InputSpec) DeepCopy() *InputSpec {
	if in == nil {
		return nil
	}
	out := new(InputSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputTuningSpec) DeepCopyInto(out *InputTuningSpec) {
	*out = *in
	if in.RateLimitByContainer != nil {
		in, out := &in.RateLimitByContainer, &out.RateLimitByContainer
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RateLimitByNodePath != nil {
		in, out := &in.RateLimitByNodePath, &out.RateLimitByNodePath
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputTuningSpec.
func (in *InputTuningSpec) DeepCopy() *InputTuningSpec {
	if in == nil {
		return nil
	}
	out := new(InputTuningSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kafka) DeepCopyInto(out *Kafka) {
	*out = *in
	if in.Brokers != nil {
		in, out := &in.Brokers, &out.Brokers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kafka.
func (in *Kafka) DeepCopy() *Kafka {
	if in == nil {
		return nil
	}
	out := new(Kafka)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeAPIAudit) DeepCopyInto(out *KubeAPIAudit) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]auditv1.PolicyRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.OmitStages != nil {
		in, out := &in.OmitStages, &out.OmitStages
		*out = make([]auditv1.Stage, len(*in))
		copy(*out, *in)
	}
	if in.OmitResponseCodes != nil {
		in, out := &in.OmitResponseCodes, &out.OmitResponseCodes
		*out = new([]int)
		if **in != nil {
			in, out := *in, *out
			*out = make([]int, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeAPIAudit.
func (in *KubeAPIAudit) DeepCopy() *KubeAPIAudit {
	if in == nil {
		return nil
	}
	out := new(KubeAPIAudit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabelSelector) DeepCopyInto(out *LabelSelector) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.MatchExpressions != nil {
		in, out := &in.MatchExpressions, &out.MatchExpressions
		*out = make([]v1.LabelSelectorRequirement, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabelSelector.
func (in *LabelSelector) DeepCopy() *LabelSelector {
	if in == nil {
		return nil
	}
	out := new(LabelSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Loki) DeepCopyInto(out *Loki) {
	*out = *in
	if in.LabelKeys != nil {
		in, out := &in.LabelKeys, &out.LabelKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Loki.
func (in *Loki) DeepCopy() *Loki {
	if in == nil {
		return nil
	}
	out := new(Loki)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputSecretSpec) DeepCopyInto(out *OutputSecretSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputSecretSpec.
func (in *OutputSecretSpec) DeepCopy() *OutputSecretSpec {
	if in == nil {
		return nil
	}
	out := new(OutputSecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputSpec) DeepCopyInto(out *OutputSpec) {
	*out = *in
	in.OutputTypeSpec.DeepCopyInto(&out.OutputTypeSpec)
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(OutputTLSSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(OutputSecretSpec)
		**out = **in
	}
	if in.Tuning != nil {
		in, out := &in.Tuning, &out.Tuning
		*out = new(OutputTuningSpec)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputSpec.
func (in *OutputSpec) DeepCopy() *OutputSpec {
	if in == nil {
		return nil
	}
	out := new(OutputSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputTLSSpec) DeepCopyInto(out *OutputTLSSpec) {
	*out = *in
	if in.TLSSecurityProfile != nil {
		in, out := &in.TLSSecurityProfile, &out.TLSSecurityProfile
		*out = new(configv1.TLSSecurityProfile)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputTLSSpec.
func (in *OutputTLSSpec) DeepCopy() *OutputTLSSpec {
	if in == nil {
		return nil
	}
	out := new(OutputTLSSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputTuningSpec) DeepCopyInto(out *OutputTuningSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputTuningSpec.
func (in *OutputTuningSpec) DeepCopy() *OutputTuningSpec {
	if in == nil {
		return nil
	}
	out := new(OutputTuningSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputTypeSpec) DeepCopyInto(out *OutputTypeSpec) {
	*out = *in
	if in.Syslog != nil {
		in, out := &in.Syslog, &out.Syslog
		*out = new(Syslog)
		**out = **in
	}
	if in.FluentdForward != nil {
		in, out := &in.FluentdForward, &out.FluentdForward
		*out = new(FluentdForward)
		**out = **in
	}
	if in.Elasticsearch != nil {
		in, out := &in.Elasticsearch, &out.Elasticsearch
		*out = new(Elasticsearch)
		**out = **in
	}
	if in.Kafka != nil {
		in, out := &in.Kafka, &out.Kafka
		*out = new(Kafka)
		(*in).DeepCopyInto(*out)
	}
	if in.Cloudwatch != nil {
		in, out := &in.Cloudwatch, &out.Cloudwatch
		*out = new(Cloudwatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Loki != nil {
		in, out := &in.Loki, &out.Loki
		*out = new(Loki)
		(*in).DeepCopyInto(*out)
	}
	if in.GoogleCloudLogging != nil {
		in, out := &in.GoogleCloudLogging, &out.GoogleCloudLogging
		*out = new(GoogleCloudLogging)
		**out = **in
	}
	if in.Splunk != nil {
		in, out := &in.Splunk, &out.Splunk
		*out = new(Splunk)
		(*in).DeepCopyInto(*out)
	}
	if in.Http != nil {
		in, out := &in.Http, &out.Http
		*out = new(Http)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputTypeSpec.
func (in *OutputTypeSpec) DeepCopy() *OutputTypeSpec {
	if in == nil {
		return nil
	}
	out := new(OutputTypeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineSpec) DeepCopyInto(out *PipelineSpec) {
	*out = *in
	if in.OutputRefs != nil {
		in, out := &in.OutputRefs, &out.OutputRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.InputRefs != nil {
		in, out := &in.InputRefs, &out.InputRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.FilterRefs != nil {
		in, out := &in.FilterRefs, &out.FilterRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineSpec.
func (in *PipelineSpec) DeepCopy() *PipelineSpec {
	if in == nil {
		return nil
	}
	out := new(PipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReceiverSpec) DeepCopyInto(out *ReceiverSpec) {
	*out = *in
	if in.ReceiverTypeSpec != nil {
		in, out := &in.ReceiverTypeSpec, &out.ReceiverTypeSpec
		*out = new(ReceiverTypeSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReceiverSpec.
func (in *ReceiverSpec) DeepCopy() *ReceiverSpec {
	if in == nil {
		return nil
	}
	out := new(ReceiverSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReceiverTypeSpec) DeepCopyInto(out *ReceiverTypeSpec) {
	*out = *in
	if in.HTTP != nil {
		in, out := &in.HTTP, &out.HTTP
		*out = new(HTTPReceiver)
		**out = **in
	}
	if in.Syslog != nil {
		in, out := &in.Syslog, &out.Syslog
		*out = new(SyslogReceiver)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReceiverTypeSpec.
func (in *ReceiverTypeSpec) DeepCopy() *ReceiverTypeSpec {
	if in == nil {
		return nil
	}
	out := new(ReceiverTypeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Splunk) DeepCopyInto(out *Splunk) {
	*out = *in
	if in.Fields != nil {
		in, out := &in.Fields, &out.Fields
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Splunk.
func (in *Splunk) DeepCopy() *Splunk {
	if in == nil {
		return nil
	}
	out := new(Splunk)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Syslog) DeepCopyInto(out *Syslog) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Syslog.
func (in *Syslog) DeepCopy() *Syslog {
	if in == nil {
		return nil
	}
	out := new(Syslog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyslogReceiver) DeepCopyInto(out *SyslogReceiver) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyslogReceiver.
func (in *SyslogReceiver) DeepCopy() *SyslogReceiver {
	if in == nil {
		return nil
	}
	out := new(SyslogReceiver)
	in.DeepCopyInto(out)
	return out
}

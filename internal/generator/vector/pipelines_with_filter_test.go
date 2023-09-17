//go:build vector

package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("Testing Config Generation", func() {

	var f = func(clspec loggingv1.CollectionSpec, secrets map[string]*corev1.Secret, clfspec loggingv1.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			Pipelines(&clfspec, op),
		)
	}

	// Make sure it doesn't interfere with non-audit logs.

	DescribeTable("Pipeline(s) with Filter(s)", helpers.TestGenerateConfWith(f),
		Entry("Kube API Server Audit Policy", helpers.ConfGenerateTest{
			CLFSpec: loggingv1.ClusterLogForwarderSpec{
				Filters: []loggingv1.FilterSpec{{
					Name: "my-audit",
					Type: loggingv1.FilterKubeAPIAudit,
					FilterTypeSpec: loggingv1.FilterTypeSpec{
						KubeAPIAudit: &loggingv1.KubeAPIAudit{
							Rules: []auditv1.PolicyRule{
								{Level: auditv1.LevelRequestResponse, Users: []string{"*apiserver"}}, // Keep full event for user ending in *apiserver
								{Level: auditv1.LevelNone, Verbs: []string{"get"}},                   // Drop other GET requests
								{Level: auditv1.LevelMetadata},                                       // Metadata for everything else.
							}}}}},
				Pipelines: []loggingv1.PipelineSpec{
					{
						InputRefs:  []string{loggingv1.InputNameAudit},
						FilterRefs: []string{"my-audit"},
						OutputRefs: []string{loggingv1.OutputTypeLoki},
					},
				},
			},
			ExpectedConf: `
[transforms._user_defined]
type = "remap"
inputs = ["audit"]
source = '''
	if is_object(.) && .kind == "Event" && .apiVersion == "audit.k8s.io/v1" {
		res = if is_null(.objectRef.resource) { "" } else { string!(.objectRef.resource) }
		sub = if is_null(.objectRef.subresource) { "" } else { string!(.objectRef.subresource) }
		namespace = if is_null(.objectRef.namespace) { "" } else { string!(.objectRef.namespace) }
		username = if is_null(.user.username) { "" } else { string!(.user.username) }
		if sub != "" { res = res + "/" + sub }
		if includes([404,409,422,429], .responseStatus.code) { # Omit by response code.
			.level = "None"
		} else if (username != "" && match(username, r'^(.*apiserver)$') && true) {
			.level = "RequestResponse"
		} else if (includes(["get"], .verb) && true) {
			.level = "None"
		} else if (true) {
			.level = "Metadata"
		} else {
			# No rule matched, apply default rules for system events.
			if match(username, r'system:.*') { # System events
				readonly = r'get|list|watch|head|options'
				if match(string!(.verb), readonly) {
		.level = "None" # Drop read-only system events.
				} else if ((int(.responseStatus.code) < 300 ?? true) && starts_with(username, "system:serviceaccount:"+namespace)) {
		.level = "None" # Drop write events by service account for same namespace as resource or for non-namespaced resource.
				}
				if .level == "RequestResponse" {
		.level = "Request" # Downgrade RequestResponse system events.
				}
			}
		}
		# Update the event
		if .level == "None" {
			abort
		} else {
			if .level == "Metadata" {
				del(.responseObject)
				del(.requestObject)
			} else if .level == "Request" {
				del(.responseObject)
			}
		}
	}
'''
`,
		}),
	)
})

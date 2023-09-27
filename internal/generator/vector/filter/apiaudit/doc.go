// Package apiaudit 'compiles' a Kube-API server audit policy into a Vector tranform that implements the policy.
//
// This is a re-implementation of splunk-audit-exporter in Vector Remap Language.
//
// The following features are not (yet) implemented:
//   - Redactions (TODO)
//   - Drop duplicate update events (TODO)
//   - Drop events before a start time (CLO future).
//   - Adding annotations (CLO future)
//   - Metrics (Separate process)
//
// See also:
//
// - Enhancement Proposal: https://github.com/openshift/enhancements/blob/master/enhancements/kube-apiserver/audit-policy.md
// - Go type definitions of [auditv1.Policy] and [auditv1.Event]
// - Auditing in Kubernetes https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
// - K8s API doc: https://kubernetes.io/docs/reference/config-api/apiserver-audit.v1/#audit-k8s-io-v1-Policy
// - Vector transforms: https://vector.dev/docs/reference/configuration/transforms/
// - Vector Remap Language: https://vector.dev/docs/reference/vrl/
//

package apiaudit

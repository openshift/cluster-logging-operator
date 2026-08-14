#!/usr/bin/env bash
#
# Manual test for protected collector ServiceAccount admission (LOG-9714 follow-up):
# ValidatingAdmissionPolicies restrict a protected collector ServiceAccount so it
# can only be used by CLO-managed collector workloads.
#
# The trust anchor is request.userInfo (the authenticated creator), NOT Pod
# metadata. This script proves that a user who copies the collector's visible
# labels/annotations/name still cannot run a Pod/Deployment as a protected SA.
#
# This is a lightweight admission test: it verifies deny/allow at CREATE only.
# The CLF may be Not Ready (no collect roles / no LokiStack) — that is expected.
#
# Prerequisites:
#   - oc logged in as cluster-admin
#   - cluster-logging-operator running with protected-SA VAP reconciliation
#
# Usage:
#   ./hack/test-protected-sa.sh
#   NS=my-test ./hack/test-protected-sa.sh
#   ./hack/test-protected-sa.sh --no-cleanup
#   ./hack/test-protected-sa.sh --cleanup-only
#
set -euo pipefail

OC="${OC:-oc}"
NS="${NS:-protected-sa-manual-test}"
OPERATOR_NS="${OPERATOR_NS:-openshift-logging}"
COLLECTOR_SA="${COLLECTOR_SA:-log-collector}"
UNPROTECTED_SA="${UNPROTECTED_SA:-plain-sa}"
RESTRICTED_SA="${RESTRICTED_SA:-restricted-user}"
CLF_NAME="${CLF_NAME:-protected-sa-test}"
CONFIGMAP="${CONFIGMAP:-clo-protected-serviceaccounts}"
CLEANUP_ONLY=false
NO_CLEANUP=false

usage() { sed -n '2,20p' "$0" | sed 's/^# \?//'; exit "${1:-0}"; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage 0 ;;
    --cleanup-only) CLEANUP_ONLY=true; shift ;;
    --no-cleanup) NO_CLEANUP=true; shift ;;
    *) echo "Unknown option: $1" >&2; usage 1 ;;
  esac
done

pass() { echo "[PASS] $*"; }
fail() { echo "[FAIL] $*" >&2; exit 1; }
info() { echo "[INFO] $*"; }

restricted_user() { echo "system:serviceaccount:${NS}:${RESTRICTED_SA}"; }

cleanup() {
  info "Cleaning up"
  "${OC}" delete clusterlogforwarder "${CLF_NAME}" -n "${NS}" --ignore-not-found >/dev/null 2>&1 || true
  "${OC}" delete ns "${NS}" --wait=true --timeout=120s --ignore-not-found >/dev/null 2>&1 || true
}

check_vap() {
  info "Checking ValidatingAdmissionPolicy resources"
  "${OC}" get validatingadmissionpolicy clo-protected-sa-pods >/dev/null
  "${OC}" get validatingadmissionpolicybinding clo-protected-sa-pods-binding >/dev/null
  "${OC}" get validatingadmissionpolicy clo-protected-sa-workloads >/dev/null
  "${OC}" get validatingadmissionpolicybinding clo-protected-sa-workloads-binding >/dev/null
  pass "both VAPs and bindings exist"
}

setup_namespace() {
  info "Setting up namespace ${NS}"
  "${OC}" create ns "${NS}" >/dev/null
  "${OC}" create sa "${COLLECTOR_SA}" -n "${NS}" >/dev/null
  "${OC}" create sa "${UNPROTECTED_SA}" -n "${NS}" >/dev/null
  "${OC}" create sa "${RESTRICTED_SA}" -n "${NS}" >/dev/null
  # Allow the restricted user to create Pods and Deployments in the namespace.
  "${OC}" create role workload-editor -n "${NS}" \
    --verb=create,get,list,delete \
    --resource=pods,deployments.apps \
    --dry-run=client -o yaml | "${OC}" apply -f - >/dev/null
  "${OC}" create rolebinding restricted-user-workload-editor -n "${NS}" \
    --role=workload-editor \
    --serviceaccount="${NS}:${RESTRICTED_SA}" \
    --dry-run=client -o yaml | "${OC}" apply -f - >/dev/null
}

create_clf_marks_sa_protected() {
  info "Creating CLF (as admin) so the operator marks ${COLLECTOR_SA} protected"
  cat <<EOF | "${OC}" create -n "${NS}" -f - >/dev/null
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: ${CLF_NAME}
spec:
  serviceAccount:
    name: ${COLLECTOR_SA}
  outputs:
  - name: test-output
    type: lokiStack
    lokiStack:
      target: { name: lokistack, namespace: openshift-logging }
      authentication: { token: { from: serviceAccount } }
  pipelines:
  - name: test-pipe
    inputRefs: [application]
    outputRefs: [test-output]
EOF
  info "Waiting for operator to add sa_${NS}_${COLLECTOR_SA} to ${OPERATOR_NS}/${CONFIGMAP}"
  # Key format is sa_<ns>_<name> (ConfigMap keys can't contain '/').
  local key="sa_${NS}_${COLLECTOR_SA}" deadline=$((SECONDS + 60))
  while true; do
    if "${OC}" get configmap "${CONFIGMAP}" -n "${OPERATOR_NS}" -o json 2>/dev/null \
        | grep -q "\"${key}\""; then
      pass "operator marked ${COLLECTOR_SA} protected"
      return 0
    fi
    (( SECONDS >= deadline )) && fail "operator did not mark SA protected within timeout (is CLO running?)"
    sleep 2
  done
}

# A Pod spec that copies the collector's visible metadata to prove spoofing fails.
spoofed_pod_yaml() {
  local sa="$1" name="$2"
  cat <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${name}
  namespace: ${NS}
  labels:
    app.kubernetes.io/name: vector
    app.kubernetes.io/instance: ${CLF_NAME}
    app.kubernetes.io/component: collector
    app.kubernetes.io/part-of: cluster-logging
    app.kubernetes.io/managed-by: cluster-logging-operator
    vector.dev/exclude: "true"
  annotations:
    target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
spec:
  serviceAccountName: ${sa}
  containers:
  - name: c
    image: registry.redhat.io/ubi9/ubi-minimal:latest
    command: ["sleep", "3600"]
EOF
}

deployment_yaml() {
  local sa="$1" name="$2"
  cat <<EOF
apiVersion: apps/v1
kind: Deployment
metadata: { name: ${name}, namespace: ${NS} }
spec:
  replicas: 1
  selector: { matchLabels: { app: ${name} } }
  template:
    metadata:
      labels:
        app: ${name}
        app.kubernetes.io/name: vector
        app.kubernetes.io/component: collector
    spec:
      serviceAccountName: ${sa}
      containers:
      - name: c
        image: registry.redhat.io/ubi9/ubi-minimal:latest
        command: ["sleep", "3600"]
EOF
}

test_pod_denied_with_protected_sa() {
  info "Test: restricted user Pod with protected SA + copied collector metadata => DENY"
  if out="$(spoofed_pod_yaml "${COLLECTOR_SA}" evil-pod | "${OC}" create --as="$(restricted_user)" -f - 2>&1)"; then
    fail "Pod create succeeded but was expected to be denied: ${out}"
  fi
  grep -qi 'protected collector ServiceAccount' <<<"${out}" \
    || fail "Pod denied but message unexpected: ${out}"
  pass "bare Pod with protected SA denied despite copied metadata"
}

test_deployment_denied_with_protected_sa() {
  info "Test: restricted user Deployment with protected SA => DENY"
  if out="$(deployment_yaml "${COLLECTOR_SA}" evil-deploy | "${OC}" create --as="$(restricted_user)" -f - 2>&1)"; then
    fail "Deployment create succeeded but was expected to be denied: ${out}"
  fi
  grep -qi 'protected collector ServiceAccount' <<<"${out}" \
    || fail "Deployment denied but message unexpected: ${out}"
  pass "Deployment with protected SA denied"
}

test_pod_allowed_with_unprotected_sa() {
  info "Test: restricted user Pod with UNprotected SA => ALLOW (pass-through)"
  if ! out="$(spoofed_pod_yaml "${UNPROTECTED_SA}" plain-pod | "${OC}" create --as="$(restricted_user)" -f - 2>&1)"; then
    fail "Pod with unprotected SA failed but was expected to succeed: ${out}"
  fi
  pass "Pod with unprotected SA admitted"
}

main() {
  command -v "${OC}" >/dev/null 2>&1 || fail "oc not found (set OC=...)"
  "${OC}" whoami >/dev/null 2>&1 || fail "not logged in ($OC whoami failed)"

  if [[ "${CLEANUP_ONLY}" == true ]]; then cleanup; pass "cleanup complete"; exit 0; fi

  check_vap
  setup_namespace
  create_clf_marks_sa_protected
  test_pod_denied_with_protected_sa
  test_deployment_denied_with_protected_sa
  test_pod_allowed_with_unprotected_sa

  echo
  pass "all protected-SA admission tests passed"
  info "Allow path for real collector pods is exercised by the operator itself"

  if [[ "${NO_CLEANUP}" == false ]]; then cleanup; info "namespace ${NS} removed"; else info "leaving namespace ${NS} (--no-cleanup)"; fi
}

main "$@"

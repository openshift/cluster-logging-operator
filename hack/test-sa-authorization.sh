#!/usr/bin/env bash
#
# Manual test for CVE-2026-10609 / LOG-9441:
# ValidatingAdmissionPolicy enforces 'use' on ServiceAccounts referenced in CLFs.
#
# This script is intentionally lightweight: it verifies admission only (create/update
# deny/allow). The CLF may show Not Ready / Unauthorized in status because the test
# does not bind collect-log ClusterRoles or deploy LokiStack — that is expected.
#
# Prerequisites:
#   - oc logged in to a cluster
#   - cluster-logging-operator running with SA usage VAP reconciliation enabled
#
# Usage:
#   ./hack/test-sa-authorization.sh
#   NS=my-test ./hack/test-sa-authorization.sh
#   ./hack/test-sa-authorization.sh --no-cleanup
#   ./hack/test-sa-authorization.sh --keep-clf
#   ./hack/test-sa-authorization.sh --cleanup-only
#
set -euo pipefail

OC="${OC:-oc}"
NS="${NS:-sa-auth-manual-test}"
COLLECTOR_SA="${COLLECTOR_SA:-log-collector}"
RESTRICTED_SA="${RESTRICTED_SA:-restricted-user}"
CLF_NAME="${CLF_NAME:-sa-auth-test}"
CLEANUP_ONLY=false
NO_CLEANUP=false
KEEP_CLF=false

usage() {
  sed -n '2,15p' "$0" | sed 's/^# \?//'
  exit "${1:-0}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage 0 ;;
    --cleanup-only) CLEANUP_ONLY=true; shift ;;
    --no-cleanup) NO_CLEANUP=true; shift ;;
    --keep-clf) KEEP_CLF=true; NO_CLEANUP=true; shift ;;
    *) echo "Unknown option: $1" >&2; usage 1 ;;
  esac
done

pass() { echo "[PASS] $*"; }
fail() { echo "[FAIL] $*" >&2; exit 1; }
info() { echo "[INFO] $*"; }

wait_for_namespace_gone() {
  if ! "${OC}" get ns "${NS}" >/dev/null 2>&1; then
    return 0
  fi
  info "Waiting for namespace ${NS} to finish terminating"
  local deadline=$((SECONDS + 120))
  while "${OC}" get ns "${NS}" >/dev/null 2>&1; do
    if (( SECONDS >= deadline )); then
      fail "timed out waiting for namespace ${NS} to be deleted"
    fi
    sleep 2
  done
}

wait_for_namespace_active() {
  local deadline=$((SECONDS + 60))
  while true; do
    phase="$("${OC}" get ns "${NS}" -o jsonpath='{.status.phase}' 2>/dev/null || echo "")"
    if [[ "${phase}" == "Active" ]]; then
      return 0
    fi
    if (( SECONDS >= deadline )); then
      fail "namespace ${NS} not Active (phase=${phase:-missing})"
    fi
    sleep 1
  done
}

cleanup() {
  info "Cleaning up namespace ${NS}"
  if "${OC}" get ns "${NS}" >/dev/null 2>&1; then
    "${OC}" delete ns "${NS}" --wait=true --timeout=120s
  fi
  wait_for_namespace_gone
}

restricted_user() {
  echo "system:serviceaccount:${NS}:${RESTRICTED_SA}"
}

clf_yaml() {
  cat <<EOF
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
      target:
        name: lokistack
        namespace: openshift-logging
      authentication:
        token:
          from: serviceAccount
  pipelines:
  - name: test-pipeline
    inputRefs:
    - application
    outputRefs:
    - test-output
EOF
}

oc_create_clf_as_restricted() {
  clf_yaml | "${OC}" create -n "${NS}" --as="$(restricted_user)" -f - 2>&1
}

grant_clf_access() {
  "${OC}" create role clf-editor -n "${NS}" \
    --verb=create,update,patch,get,list,watch,delete \
    --resource=clusterlogforwarders.observability.openshift.io \
    --dry-run=client -o yaml | "${OC}" apply -f -

  "${OC}" create rolebinding restricted-user-clf-editor -n "${NS}" \
    --role=clf-editor \
    --serviceaccount="${NS}:${RESTRICTED_SA}" \
    --dry-run=client -o yaml | "${OC}" apply -f -
}

grant_sa_usage() {
  cat <<EOF | "${OC}" apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: sa-user
  namespace: ${NS}
rules:
- apiGroups: [""]
  resources: ["serviceaccounts"]
  resourceNames: ["${COLLECTOR_SA}"]
  verbs: ["use"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: restricted-user-sa-user
  namespace: ${NS}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: sa-user
subjects:
- kind: ServiceAccount
  name: ${RESTRICTED_SA}
  namespace: ${NS}
EOF
}

revoke_sa_usage() {
  "${OC}" delete rolebinding restricted-user-sa-user -n "${NS}" --ignore-not-found >/dev/null 2>&1 || true
  "${OC}" delete role sa-user -n "${NS}" --ignore-not-found >/dev/null 2>&1 || true
}

check_vap() {
  info "Checking ValidatingAdmissionPolicy resources"
  "${OC}" get validatingadmissionpolicy clf-sa-usage-authorization >/dev/null
  "${OC}" get validatingadmissionpolicybinding clf-sa-usage-authorization-binding >/dev/null
  pass "VAP and binding exist"
}

setup_namespace() {
  wait_for_namespace_gone
  info "Setting up namespace ${NS}"
  "${OC}" create ns "${NS}"
  wait_for_namespace_active
  "${OC}" create sa "${COLLECTOR_SA}" -n "${NS}"
  "${OC}" create sa "${RESTRICTED_SA}" -n "${NS}"
  grant_clf_access
}

test_create_denied_without_use() {
  info "Test: create CLF without 'use' on ServiceAccount should be denied"
  revoke_sa_usage
  if out="$(oc_create_clf_as_restricted)"; then
    fail "CLF create succeeded but was expected to be denied. Output: ${out}"
  fi
  if ! grep -qi 'not authorized to use' <<<"${out}"; then
    fail "CLF create was denied but message was unexpected: ${out}"
  fi
  pass "create denied without 'use'"
}

test_create_allowed_with_use() {
  info "Test: create CLF with 'use' on ServiceAccount should succeed"
  grant_sa_usage
  if ! out="$(oc_create_clf_as_restricted 2>&1)"; then
    fail "CLF create failed but was expected to succeed: ${out}"
  fi
  pass "create allowed with 'use'"
}

test_update_denied_without_use() {
  info "Test: update CLF without 'use' should be denied"
  revoke_sa_usage
  if out="$("${OC}" patch clusterlogforwarder "${CLF_NAME}" -n "${NS}" \
    --as="$(restricted_user)" \
    --type=merge -p '{"metadata":{"annotations":{"sa-auth-test":"update-attempt"}}}' 2>&1)"; then
    fail "CLF update succeeded but was expected to be denied. Output: ${out}"
  fi
  if ! grep -qi 'not authorized to use' <<<"${out}"; then
    fail "CLF update was denied but message was unexpected: ${out}"
  fi
  pass "update denied without 'use'"
}

test_delete_allowed_without_use() {
  info "Test: delete CLF without 'use' should still succeed"
  if ! "${OC}" delete clusterlogforwarder "${CLF_NAME}" -n "${NS}" \
    --as="$(restricted_user)" --ignore-not-found >/dev/null; then
    fail "CLF delete failed but was expected to succeed"
  fi
  pass "delete allowed without 'use'"
}

main() {
  if ! command -v "${OC}" >/dev/null 2>&1; then
    fail "oc not found in PATH (set OC=... if needed)"
  fi
  if ! "${OC}" whoami >/dev/null 2>&1; then
    fail "not logged in to cluster (${OC} whoami failed)"
  fi

  if [[ "${CLEANUP_ONLY}" == true ]]; then
    cleanup || true
    pass "cleanup complete"
    exit 0
  fi

  check_vap
  setup_namespace
  test_create_denied_without_use
  test_create_allowed_with_use
  test_update_denied_without_use
  if [[ "${KEEP_CLF}" == false ]]; then
    test_delete_allowed_without_use
  else
    info "Skipping delete test (--keep-clf); CLF ${CLF_NAME} remains in ${NS}"
    "${OC}" get clusterlogforwarder "${CLF_NAME}" -n "${NS}" >/dev/null
    pass "CLF ${CLF_NAME} exists in namespace ${NS}"
  fi

  echo
  pass "all manual SA authorization tests passed"
  info "CLF may be Not Ready afterward — that is expected for this lightweight admission test"

  if [[ "${NO_CLEANUP}" == false ]]; then
    cleanup
    info "namespace ${NS} removed"
  else
    info "leaving namespace ${NS} (--no-cleanup)"
  fi
}

main "$@"

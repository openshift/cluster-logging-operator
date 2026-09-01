# Protecting collector ServiceAccounts from reuse by arbitrary workloads

Status: design note / investigation (prototype implemented)
Scope: Cluster Logging Operator (observability.openshift.io/v1 `ClusterLogForwarder`)

> New to ValidatingAdmissionPolicy? Read `docs/design/validatingadmissionpolicy-guide.md`
> first — it explains VAP mechanics (Policy / Binding / param) and reads our
> manifests line by line. This note covers the *why* (threat model); the guide
> covers the *how*.

Key files:
`internal/admission/protected_sa_policy.go`, `internal/admission/protected-sa-{pods,workloads}{,-binding}.yaml`,
`internal/controller/admission/protected_sa_controller.go` / `_runnable.go`;
manual test `hack/test-protected-sa.sh`. Validated end-to-end against a live
OpenShift API server via `hack/test-protected-sa.sh`: a restricted user is
denied creating a Pod and a Deployment as a protected collector SA even when
copying the collector's labels/annotations/name, while an unprotected SA passes
through. ConfigMap param keys use the form `sa_<ns>_<name>` ('/' is not a valid
ConfigMap data key).

## TL;DR

A user who can write a `ClusterLogForwarder` (CLF) can name **any** existing
ServiceAccount (SA) in the CLF namespace as the collector SA. That SA is often
bound to `logging-scc` (hostPath mounts, RunAsAny UID, SELinux `spc_t`) and to
cluster-wide read of pods/namespaces/nodes. Nothing today stops the same user
from running an **arbitrary** Pod/Deployment as that SA and inheriting those
privileges.

**A ValidatingAdmissionPolicy (VAP) can enforce the boundary, but only if it
keys on `request.userInfo` (the authenticated creating identity), which the API
server sets and a requester cannot forge. A VAP keyed on Pod labels /
annotations / names / ownerReferences would be security theater — all of that
metadata is fully reproducible by any user who can create Pods in the
namespace.**

An alternative was considered and rejected (CVE-2026-10609, LOG-9714/LOG-9441): a
VAP on the **CLF resource** requiring the CLF author to hold the `use` verb on the
referenced SA. It fails the upgrade constraint because `use` on `serviceaccounts`
is a non-standard verb **nobody holds by default**, so every existing CLF author
would be denied on upgrade until an admin provisions new `use` Roles — i.e. it
"requires changes in SA" RBAC, the exact upgrade breakage we must avoid. The
approach below instead moves enforcement to Pod admission and keys on the
operator's identity, so **no user or SA RBAC changes are required**. See §6 for
the full comparison.

The only pre-existing control is a reconcile-time `SubjectAccessReview` that
checks whether the *collector SA* has the `collect` verb — it never checks whether
the *CLF author* is allowed to use the SA.

---

## 0. Scope of this change (decision)

The original threat has **one root cause** (a CLF author may name any existing
SA) and **three exploit paths**:

1. **Reuse the SA in another pod → mount host FS → node access** (via the SCC the
   SA is bound to).
2. **Forward logs the author cannot access** — reference the SA in a CLF; the
   operator deploys a collector *as* that SA and forwards to an output the author
   controls.
3. **Exfiltrate a token** — a token-forwarding output (`BearerToken.From:
   serviceAccount`) ships the SA's bearer token to an attacker-controlled URL.

**What this change ships: Path 1 only** — the protected-SA Pod/workload VAP
(§3). It is enforced (`Deny`) with **zero upgrade breakage**, because the only
legitimate creator of a protected-SA pod is a stable, known identity (the
operator SA / the built-in controllers), which we allow-list. Existing
collectors keep running; only *new, non-operator* workloads that reuse a
protected SA are denied.

**What is deferred: paths 2 and 3 (the CLF-layer control).** They are real, but
they differ structurally from path 1: the legitimate actor is an **arbitrary
human CLF author** whose permissions cannot be predicted, so there is **no
stable identity to allow-list**. Any admission `Deny` that adds an author→SA
requirement can therefore deny *some* existing author on upgrade. Since
**no-existing-author-breakage is a hard requirement**, a hard `Deny` for
paths 2/3 cannot ship today.

> Note: the CLF `use`-verb VAP (§6) is the canonical example of the breakage we
> must avoid — `use` on `serviceaccounts` is non-default, so *every* existing
> author would be denied on upgrade until an admin grants a new Role. It is
> **not** being pursued as-is.

### Recommended future work for paths 2/3 (not implemented here)

A CLF-admission VAP keyed on `request.userInfo` can close paths 2/3 **without**
breaking existing authors if it is **surgically scoped** so it never touches an
existing CLF:

- Gate **CREATE** of a CLF, and **UPDATE only when `spec.serviceAccount.name`
  changes** (`object.spec.serviceAccount.name != oldObject.spec.serviceAccount.name`).
  Existing CLFs and same-SA edits are never re-admitted, so no existing author is
  denied on upgrade.
- Test an author→SA relationship the legitimate owner already holds (candidates:
  *can manage the SA* — `use`/`update`/`patch`/`delete` on that
  `serviceaccounts` object — or *capability parity* — the author can `collect`
  the forwarded inputs). "Can manage the SA" closes paths 2 and 3 at their common
  root (referencing a foreign privileged SA); "capability parity" targets path 2
  specifically.
- Alternatively, ship the same check as `[Warn, Audit]` first (denies nobody,
  surfaces which CLFs *would* be blocked), then flip to `Deny` in a later release
  once the fleet is provisioned.

Token side channels (reading the `<sa>-token` Secret, `serviceaccounts/token`,
`pods/exec` into the collector) remain RBAC concerns outside admission — see §4.

---

## 1. Current behavior

### 1.1 How collector ServiceAccounts are created and used

- The collector SA is **not** created by CLO in the observability path. The
  administrator pre-creates it; the CLF references it **by name** through the
  required field `spec.serviceAccount.name`
  (`api/observability/v1/clusterlogforwarder_types.go:78-88`).
- `factory.ResourceNames` sets `ServiceAccount = clf.Spec.ServiceAccount.Name`
  verbatim (`internal/factory/resource_names.go:45`) and wires it into the pod
  spec at `internal/collector/collector.go:158`
  (`ServiceAccountName: f.ResourceNames.ServiceAccount`).
- CLO only **Gets** the SA (`internal/controller/observability/collector.go:52`);
  it then **binds RBAC/SCC to it**:
  - `use` on the `logging-scc` SCC via a namespaced Role+RoleBinding
    (`internal/auth/rbac.go:108-142`; SCC defined in
    `internal/auth/securitycontextconstraint.go:33-52` —
    `AllowHostDirVolumePlugin=true`, `RunAsUser=RunAsAny`,
    `SELinuxContext=RunAsAny`).
  - `metadata-reader` ClusterRoleBinding → cluster-wide get/list/watch on
    pods, namespaces, nodes (`internal/auth/rbac.go:90-106`).
  - `system:auth-delegator` ClusterRoleBinding (TokenReview/SAR)
    (`internal/auth/rbac.go:73-87`).
- The admin is additionally expected to bind the `collect-{application,
  infrastructure,audit}-logs` ClusterRoles (the `collect` verb on `logs`).
- Token exposure: a projected SA token is mounted at
  `/var/run/ocp-collector/serviceaccount` (1h expiry,
  `internal/collector/collector.go:349-368`). When an output uses
  `BearerToken.From: serviceAccount`, a **long-lived**
  `kubernetes.io/service-account-token` Secret named `<sa>-token` is created
  and its token is sent to the output as `Authorization: Bearer`
  (`internal/controller/observability/collector.go:49-64`,
  `internal/generator/vector/output/common/auth.go:28-34`).

### 1.2 How collector workloads are identified today

Collector pods are created from a DaemonSet (or Deployment) named exactly
`clf.Name`, in the **CLF's own namespace** (not a reserved namespace).
Identifying metadata:

| Attribute | Value | CLO-controlled | Spoofable by a namespace user |
|---|---|---|---|
| `app.kubernetes.io/name` | `vector` (constant) | yes | **yes** |
| `app.kubernetes.io/instance` | `clf.Name` | yes | **yes** (predictable) |
| `app.kubernetes.io/component` | `collector` (constant) | yes | **yes** |
| `app.kubernetes.io/part-of` / `managed-by` | constants | yes | **yes** |
| `vector.dev/exclude` | `true` (constant) | yes | **yes** |
| annotations (`secret-hash`, `configmap-hash`, workload-mgmt) | constants / content hashes | yes | **yes** (any value settable) |
| resource names | `clf.Name[-suffix]`, no random component | yes | **yes** (fully predictable) |
| `serviceAccountName` | `clf.Spec.ServiceAccount.Name` | user-supplied | **yes** |
| ownerReference **on the DaemonSet** | CLF CR, `UID=clf.UID`, controller=true | yes | UID **not** forgeable |
| ownerReference **on the Pod** | set by kube DaemonSet/ReplicaSet controller | no (kube) | user's standalone pod has different/no owner |

The **only** non-forgeable attribute is the `UID` in the ownerReference of the
**DaemonSet/Deployment**. That is not present on the Pods, and CEL in a VAP
cannot dereference it to a live object. Conclusion: **no Pod-level metadata is a
trustworthy identity signal.** (Sources:
`internal/runtime/runtime.go:114-134`, `internal/collector/collector.go:119-158`,
`internal/factory/resource_names.go:35-51`,
`internal/factory/daemonset.go:13-27`, `internal/utils/utils.go:37-49`.)

### 1.3 How the existing CLF authorization works

`internal/validations/observability/validate_permissions.go` runs at **operator
reconcile time** (not admission) and issues SARs asking: *can
`system:serviceaccount:<ns>:<sa>` do `collect` on `logs/<input>`?* Failure sets
the `Authorized=False` status condition and tears the collector down. It checks
the **SA's** permissions, **not** the requesting user's right to use the SA, and
there is **no VAP or webhook** anywhere in the repo (only unrelated test
fixtures match `ValidatingAdmissionPolicy`).

---

## 2. Threat model

**Attack path (CLF write → SA reuse):**

1. Attacker has `create/update` on `ClusterLogForwarder` in namespace `N`
   (a namespaced, delegable permission).
2. Attacker discovers a privileged SA in `N` — e.g. a collector SA already
   bound to `logging-scc` + `metadata-reader`, or any SA they can name. SA names
   are predictable (`clf.Name`-derived) and enumerable.
3. **Even without touching a CLF**, the attacker creates their own Pod /
   Deployment / DaemonSet in `N` with `serviceAccountName: <privileged-sa>`.
4. Kubernetes admits it (no control blocks SA reuse). The attacker's pod now
   runs **as** that SA.

**Privileges obtained:**

- Via `logging-scc`: `hostPath` volumes (mount the node filesystem — read
  `/var/log`, and with RunAsAny UID/`spc_t`, broad node-level read access),
  RunAsAny UID (including 0), any SELinux context. This is a node-compromise
  primitive.
- Via `metadata-reader`: cluster-wide enumeration of pods/namespaces/nodes.
- Via `system:auth-delegator`: mint/validate TokenReviews & SARs.
- If a `<sa>-token` Secret exists (token-forwarding outputs): a **long-lived**
  bearer token for the SA, usable anywhere.

The CLF write permission is the entry point, but the actual exploit is **reusing
the SA identity for a non-collector workload** — exactly the boundary we must
enforce.

---

## 3. Proposed design

Enforce, at Pod admission, the invariant:

> A Pod that uses a **protected** collector SA may exist only if it is part of a
> workload tree **rooted in an object created by the CLO operator's own
> ServiceAccount.**

The trust anchor is `request.userInfo.username` — the authenticated identity of
the API caller, set by the API server, **not** spoofable by the requester and
**not** derivable from any object field. This is what makes the boundary real.

### 3.1 Identifying protected ServiceAccounts

CEL in a VAP cannot fetch the SA object, so a per-SA label is not readable at
admission. Use an **operator-maintained param object** instead:

- CLO reconciles a ConfigMap `clo-protected-serviceaccounts` in the operator
  namespace whose `data` keys are `"sa_<namespace>_<sa-name>"` for every SA
  referenced by any CLF. The `sa_` prefix + `_` separators avoid the `/`
  character, which is illegal in ConfigMap data keys (`[-._a-zA-Z0-9]+`);
  since namespace and SA names both forbid `_`, the encoding is collision-free.
- The VAP references this ConfigMap via `paramRef`. CEL rebuilds the same key
  (`'sa_' + request.namespace + '_' + variables.sa`) and tests membership with
  a `has(params.data)` guard so an empty ConfigMap never causes a CEL error.
- The ConfigMap is **always present** (created empty at startup) so CEL never
  errors on a missing param — see §7 bootstrap.

This is explicit, operator-controlled, multi-CLF / multi-namespace aware, and
strictly stronger than a naming convention. (A naming convention is *also*
spoofable input, so it cannot be the trust basis.)

### 3.2 Identifying legitimate collector Pods

Not by metadata — by **who created the object**:

- The **only** non-controller identity permitted to create a protected-SA
  workload is the operator SA
  `system:serviceaccount:<operator-ns>:cluster-logging-operator`
  (templated from the operator's own namespace).
- Built-in kube controllers propagate from an admitted root:
  `deployment-controller` (Deployment→ReplicaSet), `replicaset-controller`
  and `daemon-set-controller` (→Pods), all in `kube-system`.

Inductive soundness: a controller only creates a child if its parent was
admitted. The parent chain always terminates at a DaemonSet/Deployment, which
only the operator SA may create. So allowing the controller SAs cannot admit a
user-rooted tree — the user's root object is denied before any controller runs.

### 3.3 The admission rule

Two policies (kept separate for clarity):

- **Policy A — Pods:** if `spec.serviceAccountName` is protected, deny unless the
  creator is `daemon-set-controller` or `replicaset-controller`.
- **Policy B — workload controllers** (`apps`: daemonsets, deployments,
  replicasets, statefulsets; `batch`: jobs, cronjobs; core:
  replicationcontrollers): if the pod template references a protected SA, deny
  unless the creator is the operator SA or `deployment-controller`.

| Case | Result |
|---|---|
| protected SA + operator-created DaemonSet/Deployment | allow (B) |
| protected SA + controller-created Pod (from admitted workload) | allow (A) |
| protected SA + user bare Pod (any copied metadata) | deny (A) |
| protected SA + user Deployment/DaemonSet/Job | deny (B) |
| unprotected SA + any Pod | pass through |

---

## 4. Bypass analysis

| Bypass attempt | Outcome | Why |
|---|---|---|
| Create a bare Pod with the protected SA | **denied** (A) | creator is the user, not a controller SA |
| Copy all collector labels/annotations/name onto the Pod | **denied** (A) | policy ignores metadata; keys on `request.userInfo` |
| Create a Deployment / DaemonSet / StatefulSet with the protected SA | **denied** (B) | creator ≠ operator SA |
| Create a Job/CronJob with the protected SA | **denied** (B) | creator ≠ operator SA; job-controller never runs |
| Forge an ownerReference to the real CLF/DaemonSet | **denied** | policy never trusts ownerRefs; a forged UID is rejected/GC'd by kube anyway |
| Modify an existing Pod to switch to the protected SA | **denied / impossible** | `serviceAccountName` is immutable on Pod UPDATE; UPDATE is also matched |
| Use another controller (custom operator) to create the Pod | **denied** (A) | that controller's SA is not in the allowlist |
| Run a look-alike ReplicaSet directly | **denied** (B) | the *user* is not the operator SA; deployment-controller is allowed but the user is not |

**Residual paths the Pod VAP does NOT cover** (out of scope for pod admission —
governed by RBAC on the SA, not by running a workload):

1. **Long-lived `<sa>-token` Secret** — a user with `get secret` in the
   namespace reads the token directly, no pod needed.
2. **TokenRequest** — a user with `create` on `serviceaccounts/token` for the SA
   mints a token directly.
3. **`pods/exec`** into the running collector pod → reach the mounted projected
   token.

These must be addressed separately (restrict `get` on the token Secret / prefer
projected tokens / restrict `serviceaccounts/token` and `pods/exec`). They do
not weaken the primary boundary (no *new workload* can assume the SA identity),
but the design note must state them honestly.

**Robustness verdict:** because the policy keys on the non-spoofable authenticated
creator identity and never on object metadata, it satisfies the primary success
criterion — a CLF-writer who knows the SA name and reproduces every visible
collector attribute still cannot run an arbitrary workload as the protected SA.

---

## 5. VAP / CEL PoC

Param (reconciled by CLO; always present):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clo-protected-serviceaccounts
  namespace: openshift-logging      # operator namespace
data:
  "sa_app-logging_collector-sa": "" # one key per sa_<clf-ns>_<sa-name>
  podCreators: "system:serviceaccount:kube-system:daemon-set-controller,system:serviceaccount:kube-system:replicaset-controller"
  workloadCreators: "system:serviceaccount:openshift-logging:cluster-logging-operator,system:serviceaccount:kube-system:deployment-controller"
```

Policy A — Pods (see `internal/admission/protected-sa-pods.yaml`):

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: clo-protected-sa-pods
spec:
  failurePolicy: Fail
  paramKind: { apiVersion: v1, kind: ConfigMap }
  matchConstraints:
    resourceRules:
      - apiGroups: [""]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["pods"]
  variables:
    - name: sa
      expression: "has(object.spec.serviceAccountName) ? object.spec.serviceAccountName : 'default'"
    - name: saKey
      expression: "'sa_' + request.namespace + '_' + variables.sa"
    - name: isProtected
      expression: "has(params.data) && (variables.saKey in params.data)"
    - name: allowedCreators
      expression: "(has(params.data) && ('podCreators' in params.data)) ? params.data['podCreators'].split(',') : []"
  validations:
    - expression: "!variables.isProtected || (request.userInfo.username in variables.allowedCreators)"
      messageExpression: "'Pod uses protected ServiceAccount \"' + request.namespace + '/' + variables.sa + '\" which is only allowed for use by authorized ClusterLogForwarders'"
      reason: Forbidden
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: clo-protected-sa-pods-binding
spec:
  policyName: clo-protected-sa-pods
  validationActions: [Deny]
  paramRef:
    name: clo-protected-serviceaccounts
    namespace: openshift-logging     # overridden at reconcile time
    parameterNotFoundAction: Allow   # fail open on missing param to avoid operator self-lockout
  matchResources: {}
```

Policy B — workload controllers (see `internal/admission/protected-sa-workloads.yaml`):

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: clo-protected-sa-workloads
spec:
  failurePolicy: Fail
  paramKind: { apiVersion: v1, kind: ConfigMap }
  matchConstraints:
    resourceRules:
      - apiGroups: ["apps"]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["daemonsets", "deployments", "replicasets", "statefulsets"]
      - apiGroups: ["batch"]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["jobs", "cronjobs"]
      - apiGroups: [""]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["replicationcontrollers"]
  variables:
    - name: podSpec
      expression: >-
        request.kind.kind == 'CronJob'
          ? object.spec.jobTemplate.spec.template.spec
          : object.spec.template.spec
    - name: sa
      expression: "has(variables.podSpec.serviceAccountName) ? variables.podSpec.serviceAccountName : 'default'"
    - name: saKey
      expression: "'sa_' + request.namespace + '_' + variables.sa"
    - name: isProtected
      expression: "has(params.data) && (variables.saKey in params.data)"
    - name: allowedCreators
      expression: "(has(params.data) && ('workloadCreators' in params.data)) ? params.data['workloadCreators'].split(',') : []"
  validations:
    - expression: "!variables.isProtected || (request.userInfo.username in variables.allowedCreators)"
      messageExpression: "'Workload uses protected ServiceAccount \"' + request.namespace + '/' + variables.sa + '\" which is only allowed for use by authorized ClusterLogForwarders'"
      reason: Forbidden
```

(Bind Policy B the same way — see `internal/admission/protected-sa-workloads-binding.yaml`.
The operator templates its own namespace into the binding's `paramRef.namespace`
at reconcile time.)

Representative test cases:

- ALLOW: operator SA creates DaemonSet `app-logging` with `serviceAccountName:
  collector-sa` (protected) → B allows; daemon-set-controller creates the Pod →
  A allows.
- ALLOW: operator SA creates a Deployment → deployment-controller creates the
  ReplicaSet (B allows) → replicaset-controller creates the Pod (A allows).
- DENY: `alice` creates a bare Pod with `collector-sa` and every collector label
  copied → A denies.
- DENY: `alice` creates a DaemonSet/Deployment/Job with `collector-sa` → B denies.
- PASS: `alice` creates a Pod with `my-own-sa` (not in param) → not protected,
  admitted normally.

---

## 6. Alternative considered: a CLF-author `use`-verb VAP

An earlier direction was a VAP on the **CLF resource** (CREATE/UPDATE) requiring
the CLF author to hold the `use` verb on the referenced SA:

```
!has(object.spec.serviceAccount) || !has(object.spec.serviceAccount.name) ||
object.spec.serviceAccount.name == '' ||
authorizer.group('').resource('serviceaccounts')
  .namespace(object.metadata.namespace).name(object.spec.serviceAccount.name)
  .check('use').allowed()
```

1. **What it would prevent:** a user referencing, *in a CLF*, an SA they are not
   authorized to `use` — including the token-forwarding vector (CLO deploys a
   collector as that SA and can forward its token to an output).
2. **What it fails to prevent:** SA reuse **outside** a CLF. It only guards the
   `clusterlogforwarders` resource; a user could still create a bare
   Pod/Deployment with the protected SA and get the SCC/RBAC privileges. The
   broadened threat (arbitrary workload → node access) is **not** covered.
3. **Why it is not shipped:** the `use`-verb requirement forces admins to grant a
   new, non-default verb to every CLF author, breaking existing installs on
   upgrade (the "changes in SA" problem). This violates the hard upgrade
   constraint. The Pod/workload VAP (this design) closes the broader and more
   severe vector with **no** user/SA RBAC change, so it is the primary fix.
4. **If the token-forwarding-via-CLF vector must also be closed later,** prefer
   doing it **without** a per-user `use` grant — e.g. constrain *which* SAs a CLF
   may reference (a namespace-local SA the operator itself provisions/labels), or
   gate token-forwarding outputs specifically, rather than requiring users to
   obtain `use`. See §0 "Recommended future work for paths 2/3". Treat it as a
   separate, explicitly-scoped decision.

Also **retain** the reconcile-time SAR (`ValidatePermissions`): it verifies the
collector SA can actually `collect` and drives the `Authorized` status. It is
orthogonal to the VAP (validates SA capability, not who may run as the SA).

Net: **the Pod/workload VAP is the primary mitigation**; the SAR is **retained**;
the CLF-author `use`-verb approach is **not pursued** (its RBAC-grant upgrade cost
is the blocker), and any future CLF-layer control must need no new per-user grants.

---

## 7. Implementation recommendation

**Components to change (all operator-side; no collector SA/SCC changes):**

- New reconciler that maintains the param ConfigMap: on CLF add/update/delete,
  recompute the set of `<ns>/<sa>` keys. Likely alongside
  `internal/controller/observability/` and `internal/auth/`.
- New reconciler that ensures the two VAPs + bindings exist and are self-healing
  (reuse the `internal/reconcile` CreateOrUpdate pattern). Template the operator
  namespace into the operator-SA username via the downward API / existing
  `OPERATOR_NAME` env.
- Ship the VAP/param as **operator-reconciled objects**, not static OLM bundle
  manifests, so the operator-SA username and param stay in sync and self-heal.

**RBAC:**

- **No change to collector SAs, their RBAC, or SCC bindings** — satisfies the
  "no user migration / no new SA permissions" constraint.
- The **operator's own** ClusterRole needs `admissionregistration.k8s.io`
  (`validatingadmissionpolicies`, `validatingadmissionpolicybindings`:
  create/update/patch/delete/get/list/watch) plus create/update on the param
  ConfigMap. This ships via the CSV and is granted automatically on upgrade — it
  is an operator permission, not a user- or collector-facing one.

**Bootstrap / upgrade behavior:**

- Fresh install: operator creates the (empty) param first, then the policies.
  With the param always present, `failurePolicy: Fail` is safe (CEL never errors
  on a missing param).
- Upgrade: operator gains RBAC via the CSV, reconciles param + policies;
  existing collectors keep running because the operator SA is allowlisted and it
  owns them.
- Reconcile / CLF create-update / rollout: operator-created and
  controller-propagated → always allowed.
- Policy temporarily missing (before first reconcile): fail-open window during
  which SA reuse is briefly unguarded; closes as soon as the operator reconciles.
  Acceptable and self-correcting; the operator is never locked out of its own
  workloads.
- Self-lockout guard: keep the param ConfigMap always present, and template the
  correct operator-SA username; otherwise a wrong username in Policy B would
  block the operator from deploying collectors.

**Feasibility verdict:** the approach is technically sound and provides a real
security boundary, because it never relies on spoofable Pod metadata. The one
caveat is the token side channels in §4 (token Secret, TokenRequest,
pods/exec), which are RBAC concerns outside pod admission and should be tracked
separately.

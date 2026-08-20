# A practical guide to ValidatingAdmissionPolicy (VAP) — as used to protect collector ServiceAccounts

Audience: CLO developers new to ValidatingAdmissionPolicy. This explains what a
VAP is, how each piece is configured, and then walks through the *actual*
policies this operator ships to stop a protected collector ServiceAccount (SA)
from being reused by arbitrary workloads.

Companion docs:
- `docs/design/protect-collector-serviceaccounts.md` — the *why* (threat model, design rationale).
- This file — the *how* (VAP mechanics + a line-by-line reading of our manifests).

> Scope: this ships the **protected-SA Pod/workload VAP only** (blocks reusing a
> collector SA in an arbitrary pod → node access). The CLF-layer control that
> would also block "forward logs you can't access" / "exfiltrate a token" is a
> deliberate follow-up — see §0 of the design note for why and the recommended
> future design.

Source of truth (read alongside this doc):
- `internal/admission/manifests/protected-sa-pods.yaml` + `-binding.yaml`
- `internal/admission/manifests/protected-sa-workloads.yaml` + `-binding.yaml`
- `internal/admission/protected_sa_policy.go` (reconcile + param ConfigMap)
- `internal/admission/protected_sa_controller.go` / `_runnable.go` (when it runs)

---

## 1. What problem does admission control solve?

Every write to the Kubernetes API (`CREATE`, `UPDATE`, `DELETE`, `CONNECT`)
passes through a pipeline before it is persisted to etcd:

```
client → authentication → authorization (RBAC) → admission → etcd
                                                    ├── mutating admission
                                                    └── validating admission ← VAP runs here
```

- **Authorization (RBAC)** answers *"is this identity allowed to create a Pod in
  this namespace?"* — a coarse yes/no on the verb+resource.
- **Admission** answers the finer question *"is the **content** of this specific
  object acceptable?"* — e.g. "this Pod may create a Pod, but not one that uses
  *that particular* ServiceAccount."

RBAC cannot express our rule, because our rule depends on the *combination* of
**who** is making the request and **which SA the Pod references**. That is
exactly what admission control is for.

Historically that meant writing a **webhook** (a separate HTTPS service the
API server calls out to). **ValidatingAdmissionPolicy (VAP)** is the newer,
in-process alternative: you declare the rule in **CEL** (Common Expression
Language) inside a normal Kubernetes resource, and the API server evaluates it
itself. No webhook server, no certificates, no extra pod to run or scale.

> Availability: VAP is GA since Kubernetes 1.30 (OpenShift 4.17+). On older
> clusters the API is absent; our runnable detects that and silently skips
> installation (see `isUnsupportedAdmissionPolicyAPI` in the runnable).

---

## 2. The three objects that make up a VAP

A working policy is **three** resources. Keeping them separate is the whole
design of the feature — the same policy can be bound multiple times with
different parameters.

| Object | Kind | Answers | Cluster/namespaced |
|---|---|---|---|
| **Policy** | `ValidatingAdmissionPolicy` | *What is the rule?* (the CEL logic) | cluster-scoped |
| **Binding** | `ValidatingAdmissionPolicyBinding` | *Where does it apply and what happens on failure?* | cluster-scoped |
| **Param** | any resource (here a `ConfigMap`) | *What data does the rule read?* | namespaced (or cluster) |

Think of it as a function:

```
Policy  = the function body (CEL)
Param   = the arguments passed in       (params.*)
Binding = "call the function with these args, and Deny if it returns false"
```

A Policy alone does **nothing** until a Binding activates it. This trips people
up: you can `oc get validatingadmissionpolicy` and see it installed, yet nothing
is enforced because no Binding exists (or the Binding's `validationActions` is
not `Deny`).

---

## 3. Anatomy of the Policy object

Here is `protected-sa-pods.yaml`, annotated field by field.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: clo-protected-sa-pods
spec:
  failurePolicy: Fail                         # (A)
  paramKind:                                  # (B)
    apiVersion: v1
    kind: ConfigMap
  matchConstraints:                           # (C)
    resourceRules:
      - apiGroups: [""]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["pods"]
  variables:                                  # (D)
    - name: sa
      expression: "has(object.spec.serviceAccountName) ? object.spec.serviceAccountName : 'default'"
    - name: saKey
      expression: "'sa_' + request.namespace + '_' + variables.sa"
    - name: isProtected
      expression: "has(params.data) && (variables.saKey in params.data)"
    - name: allowedCreators
      expression: "(has(params.data) && ('podCreators' in params.data)) ? params.data['podCreators'].split(',') : []"
  validations:                                # (E)
    - expression: "!variables.isProtected || (request.userInfo.username in variables.allowedCreators)"
      messageExpression: "'Pod uses protected collector ServiceAccount \"' + variables.sa + '\" and may only be created by a CLO-managed collector controller'"
      reason: Forbidden
```

### (A) `failurePolicy: Fail` — what if the rule itself errors?

If CEL evaluation *errors* (e.g. references a field that doesn't exist, or the
param is malformed), `Fail` means **the request is denied**. The alternative,
`Ignore`, would admit the request. We choose `Fail` (fail-closed) so a broken
policy can never silently let an attacker through.

> This is subtly different from "param not found" — that is handled by the
> **binding**, see §4. We fail *closed* on evaluation errors but *open* on a
> missing param ConfigMap, so the operator can never lock itself out. This
> split is deliberate.

### (B) `paramKind` — what type of parameter this policy reads

Declares that `params` (used in the CEL) is a `ConfigMap`. This only names the
*type*; the specific instance is chosen by the binding's `paramRef` (§4). Omit
`paramKind` entirely if a policy needs no external data.

### (C) `matchConstraints` — which API requests trigger this policy

The API server only evaluates the policy for requests matching these rules.
Here: `CREATE` or `UPDATE` of core `v1` `pods`. Everything else (services,
configmaps, pod *deletes*, …) skips this policy entirely.

Why `UPDATE` too? So an attacker cannot create an innocent Pod and then *edit*
it to point at the protected SA. (In practice `serviceAccountName` is immutable
on Pods, but matching UPDATE is defence-in-depth and is required for the
workload kinds where the template *is* mutable.)

### (D) `variables` — reusable sub-expressions, evaluated top to bottom

Variables keep the final rule readable and are referenced as `variables.<name>`.
Each can use the ones above it. The objects available to CEL:

| CEL binding | What it is |
|---|---|
| `object` | the incoming resource (the Pod being created) |
| `oldObject` | previous version (on UPDATE; null on CREATE) |
| `request` | admission metadata — `request.userInfo`, `request.namespace`, `request.operation`, `request.kind`, … |
| `params` | the bound param object (the ConfigMap) |
| `authorizer` | lets CEL run authorization checks (not used here) |

Our four variables:

1. **`sa`** — the Pod's ServiceAccount, defaulting to `"default"` when the field
   is absent (Kubernetes does the same). Guarding with `has(...)` avoids a
   "no such field" evaluation error.
2. **`saKey`** — builds the lookup key `sa_<namespace>_<name>`. **Note it uses
   `request.namespace`, not `object.metadata.namespace`.** `request.namespace`
   is set by the API server and always populated; `object.metadata.namespace`
   can be empty on the incoming object (the server fills it in later), so it is
   the wrong thing to trust. ⚠️ The key format is `_`-separated, not `/`-separated
   — see the gotcha in §7.
3. **`isProtected`** — is this SA in the protected set? The set lives in the
   ConfigMap's `data` keys. `has(params.data)` guards the case where the
   ConfigMap exists but has no `data` at all.
4. **`allowedCreators`** — the allow-list of usernames permitted to create a
   protected-SA Pod, read from the ConfigMap's `podCreators` key (a
   comma-separated string, split into a list). Empty list if the key is absent.

### (E) `validations` — the actual rule

`expression` must evaluate to **`true` to ALLOW**. `false` → the action in the
binding (Deny) fires. Read ours as:

> Allow **if** the SA is *not* protected, **OR** the caller's username is in the
> allow-list.

```
!variables.isProtected  ||  (request.userInfo.username in variables.allowedCreators)
```

- Not a protected SA? `!isProtected` is `true` → allowed, short-circuits. Zero
  impact on every other workload in the cluster.
- Protected SA? Then it is allowed **only** if
  `request.userInfo.username` — the authenticated caller the API server
  stamped on the request — is an allowed creator.

**This is the crux of the whole design: we key on `request.userInfo.username`,
which the requester cannot forge.** We deliberately never look at Pod labels,
annotations, names, or ownerReferences, because a user who can create Pods can
reproduce *all* of those. Identity is the only trustworthy signal.

`messageExpression` builds the denial message shown to the user (a CEL string
so it can embed the SA name); `reason: Forbidden` maps to HTTP 403.

---

## 4. Anatomy of the Binding object

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: clo-protected-sa-pods-binding
spec:
  policyName: clo-protected-sa-pods           # which policy to activate
  validationActions: [Deny]                   # what to do when a validation returns false
  paramRef:                                   # which concrete param object to pass in
    name: clo-protected-serviceaccounts
    namespace: openshift-logging              # overwritten at reconcile time (see §6)
    parameterNotFoundAction: Allow            # missing param → admit (fail OPEN)
  matchResources: {}                          # no extra narrowing beyond the policy's matchConstraints
```

Key fields:

- **`policyName`** — links this binding to the policy above.
- **`validationActions`** — what a failing validation does. Options:
  - `Deny` — reject the request (what we use).
  - `Warn` — admit but return a warning header (great for a dry-run rollout).
  - `Audit` — admit but record in the audit log.
  You can combine e.g. `[Warn, Audit]` to observe impact before switching to
  `[Deny]` — a recommended way to roll a new policy out safely.
- **`paramRef`** — points at the specific ConfigMap instance. `parameterNotFoundAction`:
  - `Allow` — if the ConfigMap is missing, **admit** the request (fail open).
  - `Deny` — if missing, reject.
  We use **`Allow`** on purpose: if the param ConfigMap were ever deleted, we do
  *not* want to brick the whole cluster's Pod creation. Combined with
  `failurePolicy: Fail` on the policy, the net behaviour is: *fail closed on a
  broken rule, fail open on a missing param.*
- **`matchResources`** — an *additional* filter on top of the policy's
  `matchConstraints` (e.g. restrict to certain namespaces via labels). Empty
  `{}` means "no extra narrowing."

---

## 5. Anatomy of the param ConfigMap

The operator maintains one ConfigMap, `clo-protected-serviceaccounts`, in its
own namespace. Example contents on a live cluster:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clo-protected-serviceaccounts
  namespace: openshift-logging
data:
  # --- protected SA membership: one key per CLF's SA, value is empty ---
  sa_app-logging_collector-sa: ""
  sa_team-b_log-collector: ""

  # --- allow-lists of creator identities, comma-separated ---
  podCreators: "system:serviceaccount:kube-system:daemon-set-controller,system:serviceaccount:kube-system:replicaset-controller"
  workloadCreators: "system:serviceaccount:openshift-logging:cluster-logging-operator,system:serviceaccount:kube-system:deployment-controller"
```

Two kinds of entries:

1. **Protected-SA keys** (`sa_<ns>_<name>: ""`) — the *set* of SAs to protect.
   Only the key matters; the value is empty. CEL tests membership with
   `variables.saKey in params.data`.
2. **Creator allow-lists** — `podCreators` and `workloadCreators`, each a
   comma-separated username list that CEL `.split(',')`s. Built by
   `setCreatorKeys()` in `protected_sa_policy.go`:
   - `podCreators` = `daemon-set-controller`, `replicaset-controller` — the
     built-in controllers that create *Pods* from an already-admitted workload.
   - `workloadCreators` = the **operator SA** + `deployment-controller` — who may
     create the top-level *workload* (Deployment→ReplicaSet, DaemonSet).

### Why this two-level structure is sound

A built-in controller only creates a child object if its parent was already
admitted. The parent chain always terminates at a Deployment/DaemonSet, which
**only the operator SA** may create (`workloadCreators`). So allowing
`daemon-set-controller`/`replicaset-controller` to create Pods cannot admit a
*user-rooted* tree — the user's root object is denied first, before any
controller runs. A user's *bare* Pod has the user as creator, not a controller,
so it is denied.

---

## 6. How the operator installs and maintains all this

Nothing above is a static OLM manifest — it is **operator-reconciled**, for two
reasons: (a) the operator-SA username in `workloadCreators` depends on the
install namespace, and (b) the protected-SA set changes as CLFs come and go.

There are two moving parts:

### 6a. One-shot install of the policies + bindings (`protected_sa_runnable.go`)

`NewProtectedSAAdmissionRunnable` runs once after the manager cache starts
(leader-elected, so only one operator pod does it). It calls
`ReconcileProtectedSAPolicies` with exponential backoff, which:

1. Ensures the param ConfigMap exists with the creator allow-lists populated
   (`ensureProtectedSAConfigMap` → `setCreatorKeys`).
2. Does an initial `SyncProtectedServiceAccounts` (best-effort).
3. Creates/updates both Policies and both Bindings from the embedded manifests
   (`//go:embed`), **overwriting each binding's `paramRef.namespace` with the
   operator's real namespace** (line: `binding.Spec.ParamRef.Namespace = operatorNS`).

If the cluster's API server doesn't support VAP, it detects the error and skips.

### 6b. Keeping the protected-SA set current (`protected_sa_controller.go`)

`ProtectedSAReconciler` is a controller-runtime reconciler watching
`ClusterLogForwarder`. On **every** CLF create/update/delete it calls
`SyncProtectedServiceAccounts`, which:

- Lists **all** CLFs cluster-wide.
- Rebuilds the set of `sa_<ns>_<name>` keys from scratch.
- Rewrites the ConfigMap `data` (membership keys + creator allow-lists).

Rebuilding from the full list (rather than incrementally adding/removing) is why
**deletion needs no finalizer** and the ConfigMap is **self-healing**: whatever
the current CLFs are, the ConfigMap converges to match.

```
CLF created/updated/deleted
        │
        ▼
ProtectedSAReconciler.Reconcile
        │
        ▼
SyncProtectedServiceAccounts ──rebuild──► clo-protected-serviceaccounts ConfigMap
                                                    │ (paramRef)
                                                    ▼
                           VAPs read it on every Pod/workload CREATE/UPDATE
```

### Operator RBAC required

The operator's own ClusterRole needs (shipped via the CSV, granted on upgrade —
these are operator permissions, not user- or collector-facing):

- `admissionregistration.k8s.io`: `validatingadmissionpolicies`,
  `validatingadmissionpolicybindings` — create/update/patch/delete/get/list/watch.
- create/update on the param ConfigMap in the operator namespace.

**No changes to collector SAs, their RBAC, or SCC** — that is the entire point:
the collector keeps working exactly as before, and only the "reuse the SA for
another workload" move is blocked (see the design note).

---

## 7. Gotchas we already hit (learn from these)

1. **ConfigMap data keys cannot contain `/`.** They must match
   `[-._a-zA-Z0-9]+`. The first design used `sa/<ns>/<name>` and every sync
   failed — and because the binding fails *open* (`parameterNotFoundAction:
   Allow`), nothing was denied, which looked like a passing test. Fixed to
   `sa_<ns>_<name>`; both namespaces and SA names forbid `_`, so the encoding is
   collision-free. **The CEL in the manifest builds the exact same key
   (`'sa_' + request.namespace + '_' + variables.sa`) — if you change the format
   in Go, change it in both manifests too.**

2. **The fake client does not validate ConfigMap key charset.** Unit tests with
   the fake client passed while the live cluster rejected the key. Validate
   VAP/param prototypes against a **real** API server (envtest or a cluster),
   not just the fake client. That is why we added the envtest suite.

3. **A newly created VAP is not enforced instantly.** The API server compiles
   and loads policies asynchronously (seconds). Tests must poll (an `Eventually`
   that keeps trying a known-bad create until it is denied) before asserting the
   deterministic cases. See the "canary" loop in the envtest.

4. **A Policy without a Binding does nothing**, and a Binding without
   `validationActions: [Deny]` only warns/audits. If "nothing is being blocked,"
   check the Binding first.

5. **Use `request.namespace`, not `object.metadata.namespace`** for the key —
   the latter can be empty on the incoming object.

6. **Redeploy the right thing.** If you change the CEL manifests, re-apply them
   (`make apply-admission` for dev, or roll the operator so it re-reconciles). If
   you change the Go reconcile logic, rebuild/redeploy the operator image. A
   stale operator with new manifests (or vice-versa) produces confusing results.

---

## 8. How to operate and debug it

Inspect what is installed:

```bash
oc get validatingadmissionpolicy clo-protected-sa-pods -o yaml
oc get validatingadmissionpolicybinding clo-protected-sa-pods-binding -o yaml
oc get configmap clo-protected-serviceaccounts -n openshift-logging -o yaml
```

Confirm a specific SA is protected (key contains a name, so match raw JSON —
jsonpath can't select keys with dots):

```bash
oc get configmap clo-protected-serviceaccounts -n openshift-logging -o json \
  | grep 'sa_<namespace>_<sa-name>'
```

Manually verify enforcement (should be denied):

```bash
oc create --as system:serviceaccount:<ns>:some-user -n <ns> -f - <<'EOF'
apiVersion: v1
kind: Pod
metadata: { name: probe }
spec:
  serviceAccountName: <protected-sa>
  containers: [{ name: c, image: registry.redhat.io/ubi9/ubi-minimal:latest }]
EOF
# → Error ... "Pod uses protected collector ServiceAccount ... may only be created by a CLO-managed collector controller"
```

Roll out safely on a new policy: set `validationActions: [Warn, Audit]` first,
watch for unexpected warnings/audit entries, then switch to `[Deny]`.

If enforcement seems off, check in this order: (1) does the Binding exist with
`[Deny]`? (2) does the ConfigMap have the `sa_<ns>_<name>` key? (3) is the caller
accidentally in `podCreators`/`workloadCreators`? (4) is the cluster new enough
for VAP?

---

## 9. Test coverage (three layers)

| Layer | File | What it proves |
|---|---|---|
| Unit (fake client) | `internal/admission/protected_sa_policy_test.go` | The generated objects (keys, creator lists, decoded manifests) are shaped correctly. Fast, no cluster. |
| **envtest** (real apiserver) | `internal/admission/protected_sa_envtest_test.go` | The **CEL is actually compiled and enforced**. Catches invalid keys / CEL regressions. Run with `make test-admission-envtest`; skips cleanly when `KUBEBUILDER_ASSETS` is unset. |
| e2e (live cluster) | `test/e2e/collection/protected_sa/` | End-to-end with `oc create --as` impersonation against a real OpenShift cluster. |

The envtest is the layer that would have caught the `/`-in-key gotcha, because
it runs a genuine kube-apiserver rather than the fake client.

---

## 10. One-paragraph summary

A **ValidatingAdmissionPolicy** lets the API server itself enforce a CEL rule at
admission time — no webhook. It is three objects: the **Policy** (the CEL rule +
which requests it matches + fail-closed behaviour), the **Binding** (activates
the policy, says `Deny`, and points at the param, failing *open* if the param is
gone), and a **param ConfigMap** the operator keeps in sync. Our rule reads
*"if the Pod uses a protected SA, allow only if the authenticated creator
(`request.userInfo.username`) is on the allow-list."* Because it keys on the
non-forgeable caller identity and never on Pod metadata, copying the collector's
labels/name/annotations does not bypass it — which is exactly the security
boundary we need, achieved without changing any user or ServiceAccount RBAC.
</content>
</invoke>

# Dependabot Configuration Guide

This document explains how Dependabot is configured for this repository and how it works.

## What is Dependabot?

Dependabot is a GitHub native tool that automatically:
- Scans dependencies for security vulnerabilities
- Creates pull requests to update dependencies
- Keeps dependencies up-to-date with minimal manual intervention

## Current Configuration

We have Dependabot configured for **Go modules only** (main module at `/`).

### Update Schedule

- **Day**: Monday
- **Time**: 09:00 AM Eastern Time
- **Frequency**: Weekly

### What Gets Updated

1. **Direct dependencies** - all packages listed in `require` section of `go.mod` (all updates)
2. **Indirect dependencies** - transitive dependencies from `go.sum` (all updates)

Security vulnerabilities in both direct and indirect dependencies are always updated, regardless of configuration.

## How Dependabot Works

### 1. Automatic PR Creation

Every Monday morning, Dependabot:
1. Scans `go.mod` and `go.sum` files
2. Checks for available updates
3. Groups related updates together (see grouping strategy below)
4. Creates PRs with:
   - Updated `go.mod` and `go.sum`
   - Changelog links
   - Release notes
   - Compatibility score

### 2. Grouping Strategy

To reduce PR noise, updates are grouped:

| Group | Pattern | Update Types | Description |
|-------|---------|--------------|-------------|
| `aws-sdk` | `github.com/aws/aws-sdk-go-v2*` | minor, patch | AWS SDK updates together |
| `k8s-ecosystem` | `k8s.io/*`, `sigs.k8s.io/*`, `github.com/openshift/*` | minor, patch | Kubernetes/OpenShift deps together |
| `testing` | `github.com/onsi/ginkgo*`, `github.com/onsi/gomega*` | minor, patch | Test framework updates together |
| `opentelemetry` | `go.opentelemetry.io/otel*` | minor, patch | OpenTelemetry deps together |
| `golang-x` | `golang.org/x/*` | minor, patch | Go extended libraries together |
| `go-openapi` | `github.com/go-openapi/*` | minor, patch | OpenAPI libraries together |
| `patch-updates` | `*` | patch | All remaining patch updates (evaluated last) |

**Major version updates** are always created as individual PRs for careful review.

### 3. Pull Request Metadata

Each Dependabot PR includes:
- **Labels**: `dependencies`, `go`, `automated`
- **Reviewers**: Team members from OWNERS file (jcantrill, vparfonov, Clee2691)
- **Assignees**: jcantrill (for tracking)
- **Commit message**: Prefixed with `chore(deps):` for consistency
- **Auto-rebase**: Enabled (see conflict resolution below)

### 4. PR Limits

Maximum of **10 open PRs** at any time to prevent overwhelming the review queue.

## Conflict Resolution

### Automatic Rebase

Dependabot has `rebase-strategy: "auto"` enabled, which means:

1. **When base branch changes**: Dependabot automatically rebases its PRs
2. **When conflicts occur**: Dependabot will:
   - Attempt to auto-rebase
   - If rebase succeeds: Updates the PR automatically
   - If rebase fails: Closes the PR and recreates it on next run

### Handling Merge Conflicts

**Scenario 1: Clean rebase**
```
Base branch updated → Dependabot rebases → PR auto-updates → Ready for merge
```

**Scenario 2: Rebase conflict**
```
Base branch updated → Dependabot rebase fails → PR closed → New PR created next Monday
```

**Manual intervention needed when:**
- Custom changes were made to dependency versions in `go.mod`
- Lock file (`go.sum`) has manual modifications
- Multiple competing dependency updates

### Manual Conflict Resolution

If you need to manually resolve conflicts in a Dependabot PR:

```bash
# 1. Checkout the PR branch
gh pr checkout <PR-number>

# 2. Rebase on master
git rebase master

# 3. Resolve conflicts in go.mod/go.sum
# Edit files as needed

# 4. Continue rebase
git add go.mod go.sum
git rebase --continue

# 5. Force push (Dependabot will detect and accept it)
git push --force-with-lease
```

**Note**: You can also comment `@dependabot rebase` on the PR to trigger a rebase.

## Dependabot Commands

You can interact with Dependabot via PR comments:

| Command | Description |
|---------|-------------|
| `@dependabot rebase` | Rebase the PR against the base branch |
| `@dependabot recreate` | Recreate the PR from scratch |
| `@dependabot merge` | Merge the PR (if CI passes) |
| `@dependabot squash and merge` | Squash and merge the PR |
| `@dependabot cancel merge` | Cancel a previous merge request |
| `@dependabot reopen` | Reopen a closed PR |
| `@dependabot close` | Close the PR and ignore future updates |
| `@dependabot ignore this dependency` | Never update this dependency again |
| `@dependabot ignore this major version` | Ignore updates to this major version |
| `@dependabot ignore this minor version` | Ignore updates to this minor version |

## Reviewing Dependabot PRs

### Quick Review Checklist

1. **Check CI status**: All tests must pass
2. **Review changelog**: Click through to release notes
3. **Check compatibility**: Look at the compatibility score
4. **Security updates**: Prioritize these
5. **Breaking changes**: Review carefully for major versions
6. **Run vulnerability scan**: `make lint-vuln` to verify fixes

### Security Updates

Security updates are labeled and should be prioritized:
- Review the CVE details in the PR description
- Verify with `govulncheck -show color,verbose ./...`
- Merge quickly after CI passes

### Testing Dependabot Updates

```bash
# 1. Checkout the PR
gh pr checkout <PR-number>

# 2. Run full test suite
make check

# 3. Run vulnerability scan
make lint-vuln

# 4. Build and test locally
make build
make test-unit
```

## Troubleshooting

### PR Keeps Getting Recreated

**Cause**: Base branch has incompatible changes

**Solution**:
1. Manually merge the PR with conflict resolution
2. Or ignore the dependency: `@dependabot ignore this dependency`

### Too Many PRs

**Cause**: Multiple dependencies have updates

**Solutions**:
1. Merge compatible PRs in batches
2. Reduce `open-pull-requests-limit` in config
3. Adjust grouping strategy to be more aggressive

### CI Fails on Dependabot PR

**Common causes**:
- Breaking API changes in dependency
- Test compatibility issues
- Transitive dependency conflicts

**Solutions**:
1. Check dependency changelog for breaking changes
2. Update calling code if needed
3. Pin to previous version if update is incompatible
4. Comment `@dependabot ignore this major version`

### Dependabot Not Creating PRs

**Check**:
1. Is Dependabot enabled for the repo? (Settings → Security → Dependabot)
2. Are there existing PRs at the limit? (default: 10)
3. Check Dependabot logs: Repository → Insights → Dependency graph → Dependabot

## Disabling Updates for Specific Dependencies

If a dependency should not be auto-updated (e.g., pinned for compatibility):

**Option 1**: Via PR comment
```
@dependabot ignore this dependency
```

**Option 2**: Via config file
```yaml
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    ignore:
      - dependency-name: "github.com/example/package"
        # Ignore all updates
      - dependency-name: "k8s.io/client-go"
        # Ignore only major version updates
        update-types: ["version-update:semver-major"]
```

## Best Practices

1. **Review weekly**: Set aside time Monday afternoons to review Dependabot PRs
2. **Security first**: Merge security updates ASAP
3. **Group merging**: Merge compatible patches together to reduce churn
4. **Test before merge**: Run `make check` on important updates
5. **Keep PRs fresh**: Don't let PRs sit too long; they'll need rebasing
6. **Monitor CI**: Set up Slack/email notifications for failed Dependabot builds

## Integration with govulncheck

Our `make lint` now includes vulnerability scanning with `govulncheck`. This complements Dependabot:

- **Dependabot**: Proactively creates PRs for vulnerable dependencies
- **govulncheck**: Verifies vulnerabilities are actually fixed

Workflow:
1. Dependabot creates PR for security update
2. CI runs `make lint` which includes `govulncheck`
3. Verify warning count decreases
4. Merge if clean

## Configuration File Location

`.github/dependabot.yml` - Edit this file to modify Dependabot behavior

## Further Reading

- [Dependabot documentation](https://docs.github.com/en/code-security/dependabot)
- [Configuration options](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
- [Go modules support](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file#package-ecosystem)

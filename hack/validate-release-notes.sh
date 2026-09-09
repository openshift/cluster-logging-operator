#!/usr/bin/env bash

# Validate that commits reference LOG-XXXX or CVE and that references are in release notes
# Checks commits since the 6.5.3 release
# Usage: validate-release-notes.sh [base-ref]
# Default base-ref: last commit with "Update version to 6.5.3"

set -euo pipefail

RELEASE_NOTES="${RELEASE_NOTES:-release-notes.adoc}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"

# Check if release notes file exists
if [[ ! -f "${RELEASE_NOTES}" ]]; then
  echo "ERROR: Release notes file not found: ${RELEASE_NOTES}"
  exit 1
fi

# If no base-ref provided, find the 6.5.3 release
if [[ -z "${1:-}" ]]; then
  LAST_6_5_3=$(git log --all --oneline | grep "Update version to 6.5.3" | head -1 | cut -d' ' -f1)
  if [[ -z "${LAST_6_5_3}" ]]; then
    echo "ERROR: Could not find 6.5.3 release"
    exit 1
  fi
  BASE_REF="${LAST_6_5_3}"
else
  BASE_REF="${1}"
fi

echo "Validating commits since: ${BASE_REF}"

FAILED=0
COMMITS=$(git rev-list "${BASE_REF}..HEAD" 2>/dev/null || echo "")

if [[ -z "${COMMITS}" ]]; then
  echo "No commits to validate against ${BASE_REF}"
  exit 0
fi

for COMMIT in ${COMMITS}; do
  MSG=$(git log -1 --pretty=format:"%B" "${COMMIT}")
  SUBJECT=$(git log -1 --pretty=format:"%s" "${COMMIT}")

  # Extract LOG-XXXX and CVE references from commit message
  REFS=$(echo "${MSG}" | grep -oE "(LOG-[0-9]+|CVE-[0-9]{4}-[0-9]+)" || true)

  if [[ -z "${REFS}" ]]; then
    echo "ERROR [${COMMIT:0:7}]: Commit does not reference a LOG-XXXX or CVE identifier"
    echo "  Subject: ${SUBJECT}"
    FAILED=1
    continue
  fi

  # Check that each reference exists in release notes
  while IFS= read -r REF; do
    if ! grep -q "${REF}" "${RELEASE_NOTES}"; then
      echo "ERROR [${COMMIT:0:7}]: Reference ${REF} not found in ${RELEASE_NOTES}"
      echo "  Subject: ${SUBJECT}"
      FAILED=1
    fi
  done <<< "${REFS}"
done

if [[ ${FAILED} -eq 1 ]]; then
  echo ""
  echo "Release notes validation failed."
  echo "Please ensure:"
  echo "  1. All commits reference a LOG-XXXX or CVE identifier"
  echo "  2. All references are documented in ${RELEASE_NOTES}"
  exit 1
fi

echo "✓ Release notes validation passed"
exit 0

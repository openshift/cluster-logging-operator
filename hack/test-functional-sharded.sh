#!/usr/bin/env bash
#
# Run the functional ginkgo suites sharded across concurrent ginkgo invocations
# so that specs AND suites run in parallel.
#
# A single `ginkgo -p` run parallelizes the specs *within* a suite but runs the
# suites themselves serially, so total wall time is the sum of the suites. This
# script splits the suites into FUNCTIONAL_SHARDS cost-balanced groups and runs
# one `ginkgo -p` invocation per group concurrently, overlapping the suites.
#
# Peak concurrency is roughly FUNCTIONAL_SHARDS * FUNCTIONAL_PROCS functional
# pods, so keep that bounded to the claimed cluster's capacity. Defaults are
# 4 shards * 4 procs = 16 peak (2x the previous single `-p --procs=8` run) since
# functional pods are light and latency-bound; dial back to e.g. 4 shards * 2
# procs if the cluster shows contention (the lesson from the lokistack e2e).
#
set -uo pipefail

SHARDS="${FUNCTIONAL_SHARDS:-4}"
PROCS="${FUNCTIONAL_PROCS:-4}"
TIMEOUT="${FUNCTIONAL_TIMEOUT:-60m}"
LOGDIR="${ARTIFACT_DIR:-$(mktemp -d)}"

mkdir -p "${LOGDIR}"

# Discover suite directories (packages that have test files) under test/functional.
mapfile -t SUITES < <(GOFLAGS=-mod=mod go list -f '{{if .TestGoFiles}}{{.Dir}}{{end}}' ./test/functional/... | sort)
if [[ ${#SUITES[@]} -eq 0 ]]; then
  echo "no functional suites found" >&2
  exit 1
fi

# Approximate per-suite wall-time hints (seconds, from an 8-proc run) so the
# bin-packer can balance shards. Suites not listed here default to a small weight
# and are spread evenly; the hints only need refreshing if a suite's cost changes
# dramatically, and getting them wrong only unbalances shards, never breaks them.
weight() {
  case "$1" in
    */outputs/elasticsearch)  echo 380 ;;
    */outputs/splunk)         echo 314 ;;
    */outputs/kafka)          echo 285 ;;
    */outputs/syslog)         echo 206 ;;
    */filters/prune)          echo 176 ;;
    */filters/drop)           echo 127 ;;
    */outputs/otlp)           echo 107 ;;
    */normalization)          echo 101 ;;
    */normalization/loglevel) echo  87 ;;
    */misc)                   echo  69 ;;
    *)                        echo  30 ;;
  esac
}

# Longest-processing-time bin-packing: walk suites heaviest-first, assigning each
# to the currently least-loaded shard.
declare -a SHARD_LOAD SHARD_DIRS
for ((i = 0; i < SHARDS; i++)); do SHARD_LOAD[i]=0; SHARD_DIRS[i]=""; done
while IFS=$'\t' read -r w dir; do
  min=0
  for ((i = 1; i < SHARDS; i++)); do
    if ((SHARD_LOAD[i] < SHARD_LOAD[min])); then min=$i; fi
  done
  SHARD_LOAD[min]=$((SHARD_LOAD[min] + w))
  SHARD_DIRS[min]="${SHARD_DIRS[min]} ${dir}"
done < <(for d in "${SUITES[@]}"; do printf '%s\t%s\n' "$(weight "$d")" "$d"; done | sort -rn)

echo "Running ${#SUITES[@]} functional suites across ${SHARDS} shards (procs=${PROCS} each); per-shard logs in ${LOGDIR}"

# Launch one ginkgo invocation per non-empty shard, concurrently. Each shard's
# output is prefixed for the live CI log and tee'd to a per-shard artifact file.
declare -a PIDS SHARD_OF_PID
for ((i = 0; i < SHARDS; i++)); do
  dirs="${SHARD_DIRS[i]# }"
  [[ -z "${dirs}" ]] && continue
  echo "  shard ${i} (~${SHARD_LOAD[i]}s weight): ${dirs}"
  (
    set -o pipefail
    # shellcheck disable=SC2086 # dirs is an intentional space-separated list
    GOFLAGS=-mod=mod go run github.com/onsi/ginkgo/v2/ginkgo \
      -race -p --procs="${PROCS}" --keep-going \
      --no-color --poll-progress-after=45s --timeout="${TIMEOUT}" \
      ${dirs} 2>&1 | tee "${LOGDIR}/functional-shard-${i}.log" | sed "s/^/[shard ${i}] /"
  ) &
  PIDS+=($!)
  SHARD_OF_PID+=("${i}")
done

# Wait for every shard; fail if any shard failed (pipefail makes each subshell's
# status the ginkgo exit code, not tee/sed).
rc=0
for idx in "${!PIDS[@]}"; do
  if ! wait "${PIDS[idx]}"; then
    echo "shard ${SHARD_OF_PID[idx]} FAILED (see ${LOGDIR}/functional-shard-${SHARD_OF_PID[idx]}.log)" >&2
    rc=1
  else
    echo "shard ${SHARD_OF_PID[idx]} passed"
  fi
done

exit "${rc}"

#!/bin/bash

current_dir=$(dirname "${BASH_SOURCE[0]}" )
GOFLAGS=-mod=mod NO_COLOR=1 go test -v -timeout=90m -ginkgo.v -ginkgo.trace \
   -ginkgo.poll-progress-after=300s \
   -ginkgo.poll-progress-interval=30s \
   "${current_dir}/../test/e2e/..."

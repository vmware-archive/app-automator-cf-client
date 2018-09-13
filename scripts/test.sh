#!/usr/bin/env bash
set -eo pipefail

pushd $(dirname $(dirname $0))
  go get github.com/onsi/ginkgo/ginkgo
  ginkgo -r -p -race -randomizeAllSpecs
popd

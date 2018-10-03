#!/usr/bin/env bash
set -eo pipefail

export GOPATH="$(cd $(dirname $0)/../../../../..; pwd -P)"

pushd "$(dirname $0)/.."
  go get github.com/onsi/ginkgo/ginkgo
  go get -t ./...
  $GOPATH/bin/ginkgo -r -p -race -randomizeAllSpecs
popd

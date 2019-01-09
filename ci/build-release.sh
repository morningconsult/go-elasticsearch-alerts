#!/bin/sh
# Copyright 2019 The Morning Consult, LLC or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#         https://www.apache.org/licenses/LICENSE-2.0
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

set -e

readonly PROJECT="github.com/morningconsult/go-elasticsearch-alerts"
readonly GORELEASER_VERSION=v0.88.0

echo "==> Installing APK dependencies"

apk add -qU --no-progress \
  make \
  bash \
  git \
  gnupg

echo "==> Installing goreleaser $GORELEASER_VERSION"

wget --quiet -O /tmp/goreleaser.tar.gz https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz
tar xzf /tmp/goreleaser.tar.gz -C /tmp
mv /tmp/goreleaser /usr/local/bin

mkdir -p "${GOPATH}/src/${PROJECT}"
cp -r . "${GOPATH}/src/${PROJECT}"
cd "${GOPATH}/src/${PROJECT}"

echo "==> Running unit tests"

CGO_ENABLED=0 make test

goreleaser release \
  --rm-dist

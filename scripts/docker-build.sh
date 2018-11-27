#!/usr/bin/env bash
# Copyright 2018 The Morning Consult, LLC or its affiliates. All Rights Reserved.
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

TOOL="go-elasticsearch-alerts"
REPO="gitlab.morningconsult.com/mci/${TOOL}"
BIN_DIR="bin/local"
ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )

cd "${ROOT}"

mkdir -p "${ROOT}/bin/local"

echo "==> Building Docker image..."

IMAGE=$( docker build \
    --quiet \
    --build-arg TARGET_GOARCH=${TARGET_GOARCH} \
    --build-arg TARGET_GOOS=${TARGET_GOOS} \
    . \
)

echo "==> Building the binary..."

CONTAINER_ID=$( docker run --rm --detach --tty ${IMAGE} )

docker cp "${CONTAINER_ID}:/go/src/${REPO}/${BIN_DIR}/${TOOL}" "${ROOT}/${BIN_DIR}"

docker kill "${CONTAINER_ID}" > /dev/null

echo "==> Done. The binary can be found in:  ${ROOT}/${BIN_DIR}/${TOOL}"
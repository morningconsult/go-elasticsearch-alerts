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

ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )
cd "${ROOT}"

BIN_DIR="${1}"
TAG="${2}"
COMMIT="${3}"
PACKAGE_ROOT="${4}"

if [ -z "${BIN_DIR}" ]; then
  echo "Directory to which binary should be written (first argument) is missing. Exiting."
  exit 1
fi

if [ -z "${PACKAGE_ROOT}" ]; then
  echo "Project name (fourth argument) is missing. This should be the project name (e.g. \"github.com/example/project\"). Exiting."
  exit 1
fi

BIN_NAME=$( basename "${PACKAGE_ROOT}" )


# Builds the binary from source in the specified destination paths.
mkdir -p "${BIN_DIR}"

cd "${ROOT}"

version_ldflags="-X \"${PACKAGE_ROOT}/version.Date=$(date +"%b %d, %Y")\""

if [[ -n "${TAG}" ]]; then
  version_ldflags="-X \"${PACKAGE_ROOT}/version.Version=${TAG}\""
fi

if [[ -n "${COMMIT}" ]]; then
  version_ldflags="${version_ldflags} -X \"${PACKAGE_ROOT}/version.Commit=${COMMIT}\""
fi

CGO_ENABLED=0 go build \
  -installsuffix cgo \
  -a \
  -ldflags "-s ${version_ldflags}" \
  -o "${BIN_DIR}/${BIN_NAME}" \
  .
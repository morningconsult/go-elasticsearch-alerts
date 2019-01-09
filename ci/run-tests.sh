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

PROJECT="github.com/morningconsult/go-elasticsearch-alerts"

# NOTE: This script is intended to be run in a
#   golang:1.11.3-alpine3.8 Docker container

# Install dependencies
apk add -qU --no-cache make git bzr

# Run tests
CGO_ENABLED=0 make test

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

FROM golang:1.11-alpine3.8

RUN set -e; \
  apk add -qU --no-cache git make; \
  rm -f /var/cache/apk/*;

ARG TARGET_GOOS
ARG TARGET_GOARCH

ENV GOOS $TARGET_GOOS
ENV GOARCH $TARGET_GOARCH

ARG BINARY=go-elasticsearch-alerts
ARG PROJECT=github.com/morningconsult/$BINARY

WORKDIR /go/src/$PROJECT

COPY . .

RUN make

ENTRYPOINT "/bin/sh"
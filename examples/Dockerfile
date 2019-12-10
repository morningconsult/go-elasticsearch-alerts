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

FROM golang:1.13.4-alpine3.10

RUN set -e; \
  apk add -qU --no-cache git bzr make;

WORKDIR /build

RUN set -e; \
  git clone https://github.com/morningconsult/go-elasticsearch-alerts; \
  cd go-elasticsearch-alerts; \
  CGO_ENABLED=0 make; \
  mv ./bin/go-elasticsearch-alerts /usr/local/bin;

COPY . .

CMD ["go-elasticsearch-alerts"]

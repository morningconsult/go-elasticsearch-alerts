
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

FLY := $(shell which fly)

REPO=gitlab.morningconsult.com/mci/go-elasticsearch-alerts
SOURCES := $(shell find . -name '*.go')
BINARY_NAME=go-elasticsearch-alerts
LOCAL_BINARY=bin/local/$(BINARY_NAME)
GOPATH := $(shell pwd)
GOBIN := $(GOPATH)/bin

all: build

update_deps:
	scripts/update-deps.sh
.PHONY: update_deps

docker: Dockerfile
	scripts/docker-build.sh
.PHONY: docker

build: $(LOCAL_BINARY)
.PHONY: build

$(LOCAL_BINARY): $(SOURCES)
	@echo "==> Starting binary build..."
	@sh -c "'./scripts/build-binary.sh' './bin/local' '$(shell git describe --tags --abbrev=0)' '$(shell git rev-parse --short HEAD)' '$(REPO)'"
	@echo "==> Done. Binary can be found at $(LOCAL_BINARY)"

test:
	@go test -v -cover $(shell go list "./..." | grep -v scripts)
.PHONY: test

git_chglog_check:
	if [ -z "$(shell which git-chglog)" ]; then \
		GOPATH=$(GOPATH) PATH=$$PATH:$(GOBIN) go get -u -v github.com/git-chglog/git-chglog/cmd/git-chglog && GOPATH=$(GOPATH) PATH=$$PATH:$(GOBIN) git-chglog --version; \
	fi
.PHONY: git_chglog_check

changelog: git_chglog_check
	GOPATH=$(GOPATH) PATH=$$PATH:$(GOBIN) git-chglog --output CHANGELOG.md
.PHONY: changelog

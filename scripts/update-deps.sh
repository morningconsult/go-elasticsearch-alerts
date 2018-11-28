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

# ORIGIN=$(pwd)
ORG="github.com/morningconsult"
TOOL="go-elasticsearch-alerts"

## Make a temporary directory
TEMPDIR=$(mktemp -d get-deps.XXXXXX)

## Set paths
export GOPATH="$(pwd)/${TEMPDIR}"
export PATH="${GOPATH}/bin:${PATH}"
cd $TEMPDIR

## Get repo
mkdir -p "src/${ORG}"
cd "src/${ORG}"
echo "Fetching ${TOOL}..."
git clone git@github.com:morningconsult/${TOOL}
cd ${TOOL}

## Clean out earlier vendoring
rm -rf Godeps vendor

## Get govendor
go get -u github.com/kardianos/govendor

govendor init

## Fetch dependencies
echo "Fetching dependencies. This will take some time..."
govendor fetch +missing

printf "Done; to commit, run: \n\n    $ cd ${GOPATH}/src/${ORG}/${TOOL}\n\n"
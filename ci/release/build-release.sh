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

set -eu

apk add -qU --no-cache --no-progress git

git config --global user.email "${GITHUB_EMAIL}"
git config --global user.name "${GITHUB_ACTOR}"

readonly GIT_CHGLOG_VERSION="0.8.0"

# Download git-chglog
wget --output-document=/usr/local/bin/git-chglog --quiet \
  "https://github.com/git-chglog/git-chglog/releases/download/${GIT_CHGLOG_VERSION}/git-chglog_linux_amd64"

# Make it executable
chmod a+x /usr/local/bin/git-chglog

# Clone the repo
git clone repo repo-dirty

cd repo-dirty

# Install dependencies
pip install --quiet --requirement ./ci/release/requirements.txt

# Get the latest version
readonly VERSION="$( python ./ci/release/version.py )"

if [ "${VERSION}" = "" ]; then
  exit 0
fi

echo "${VERSION}" > ./VERSION

git-chglog --output ./CHANGELOG.md

# Add and commit changed files
git add ./VERSION ./CHANGELOG.md
git commit -m "chore: Bump version and update changelog [ci skip]"

# Tag the repo with the latest version
git tag "v${VERSION}"

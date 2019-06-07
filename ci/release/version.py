from semantic_version import Version
from git import Repo
import os
import re

def enumerate_semver(version):
    return (int(i.strip('vV')) for i in version.split('.'))

def get_latest_version(repo):
    semver_re = re.compile("^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?(\+[0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*)?$")

    semvers = [str(tag) for tag in repo.tags if semver_re.search(str(tag))]

    semvers.sort(
        reverse=True,
        key=lambda x: tuple(enumerate_semver(x))
    )

    if len(semvers) < 1:
        return ''

    return semvers[0]

def next_version(version, repo, branch='master'):
    latest_tagged_commit = repo.commit(str(version))

    pre_v1 = version.major < 1

    new_feature = False
    for commit in repo.iter_commits(branch):
        if 'BREAKING CHANGE' in commit.summary:
            if pre_v1:
                version = version.next_minor()
            else:
                version = version.next_major()
            return version
        if commit.summary.find('feat') == 0:
            new_feature = True
        if latest_tagged_commit == commit:
            break
    if new_feature and not pre_v1:
        return version.next_minor()
    return version.next_patch()

def main():
    repo = Repo()
    latest_tag = get_latest_version(repo)
    if latest_tag == '':
        print('No semantic version tags found')
        exit(1)
    v = Version(latest_tag)
    version = next_version(v, repo)
    print(version)

if __name__ == '__main__': main()


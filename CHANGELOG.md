<a name="unreleased"></a>
## [Unreleased]


<a name="v0.1.35"></a>
## [v0.1.35] - 2021-06-17
### Chore
- Bump version and update changelog


<a name="v0.1.34"></a>
## [v0.1.34] - 2021-06-17
### Chore
- Bump version and update changelog


<a name="v0.1.33"></a>
## [v0.1.33] - 2021-04-28
### Chore
- Bump version and update changelog


<a name="v0.1.32"></a>
## [v0.1.32] - 2021-04-22
### Chore
- Bump version and update changelog


<a name="v0.1.31"></a>
## [v0.1.31] - 2021-03-10
### Chore
- Bump version and update changelog
- Remove ci skip from bump commit so release stage will run


<a name="v0.1.30"></a>
## [v0.1.30] - 2021-03-10
### Chore
- Bump version and update changelog [ci skip]
- Fix inclusify changes


<a name="v0.1.29"></a>
## [v0.1.29] - 2021-02-09
### Chore
- Bump version and update changelog [ci skip]
- Update golang versions to 1.15.8


<a name="v0.1.28"></a>
## [v0.1.28] - 2020-09-12
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.27"></a>
## [v0.1.27] - 2020-08-14
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.26"></a>
## [v0.1.26] - 2020-07-21
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.25"></a>
## [v0.1.25] - 2020-07-14
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.24"></a>
## [v0.1.24] - 2020-07-13
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.23"></a>
## [v0.1.23] - 2020-07-13
### Chore
- Bump version and update changelog [ci skip]
- remove all concourse stuff


<a name="v0.1.22"></a>
## [v0.1.22] - 2020-07-13
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.21"></a>
## [v0.1.21] - 2020-07-13
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.20"></a>
## [v0.1.20] - 2020-07-13
### Chore
- Bump version and update changelog [ci skip]

### Ci
- Testing out github workflows ([#67](https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/issues/67))


<a name="v0.1.19"></a>
## [v0.1.19] - 2020-07-13
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.18"></a>
## [v0.1.18] - 2020-02-11
### Chore
- Bump version and update changelog [ci skip]
- go mod tidy


<a name="v0.1.17"></a>
## [v0.1.17] - 2020-02-11
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.16"></a>
## [v0.1.16] - 2019-12-10
### Chore
- Bump version and update changelog [ci skip]
- Fix demonstration code
- go mod tidy

### Feat
- Allow users to make alerts fire only when certain conditions are satisfied


<a name="v0.1.15"></a>
## [v0.1.15] - 2019-10-22
### Chore
- Bump version and update changelog [ci skip]
- Add unit test for basic auth

### Feat
- Enable basic authentication for elasticsearch requests


<a name="v0.1.14"></a>
## [v0.1.14] - 2019-10-15
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.13"></a>
## [v0.1.13] - 2019-09-19
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.12"></a>
## [v0.1.12] - 2019-09-15
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.11"></a>
## [v0.1.11] - 2019-06-27
### Chore
- Bump version and update changelog [ci skip]
- Make better default alert message for SNS

### Feat
- Expose github.com/Masterminds/sprig to template


<a name="v0.1.10"></a>
## [v0.1.10] - 2019-06-24
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.9"></a>
## [v0.1.9] - 2019-06-19
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.8"></a>
## [v0.1.8] - 2019-06-19
### Chore
- Bump version and update changelog [ci skip]

### Feat
- Implement SNS output method


<a name="v0.1.7"></a>
## [v0.1.7] - 2019-06-11
### Chore
- Bump version and update changelog [ci skip]


<a name="v0.1.6"></a>
## [v0.1.6] - 2019-06-11
### Chore
- Bump version and update changelog [ci skip]
- Update deprecated field in .goreleaser.yml


<a name="v0.1.5"></a>
## [v0.1.5] - 2019-06-11
### Chore
- Bump version and update changelog [ci skip]
- Bump version and update changelog [ci skip]
- Bump version and update changelog [ci skip]
- Add golang.org/x/xerrors and enforce authentication when setting pipeline
- Fix linting issues
- Removed github.com/hashicorp/helper/jsonutil as a dependency and general housekeeping

### Ci
- Fix missing secrets [ci skip]

### Feat
- Improve pipeline


<a name="v0.1.4"></a>
## [v0.1.4] - 2019-04-13
### Ci
- FIx PR context not being interpolated
- Clean up some pipeline stuff

### Make
- Missed a line break in fly check

### Templates
- Reduce number of shards to 1


<a name="v0.1.2"></a>
## [v0.1.2] - 2019-03-26
### Chore
- Add pending step before running test-pr
- Bump Goreleaser version and use new pull request Concourse resource

### Ci
- Fix tag_filter to use glob [ci skip]


<a name="v0.1.3"></a>
## [v0.1.3] - 2019-01-10
### Chore
- Updated other files to reflect migration to go modules
- Using go modules instead of dep
- Using go modules instead of dep
- Updated license header to 2019; command/client.go: Separated logic for creating new Consul client into its own file; .drone.yml: Deleted since no longer using drone

### Ci
- fixing some errors with pipeline [ci skip]

### Dockerfile
- Update go image to 1.11.4

### Vendor
- now using go module

### Reverts
- ci: fixing some errors with pipeline [ci skip]


<a name="v0.1.1"></a>
## [v0.1.1] - 2019-01-10

<a name="v0.1.0"></a>
## [v0.1.0] - 2019-01-09
### Chore
- Using go modules instead of dep
- Using go modules instead of dep
- Updated license header to 2019; command/client.go: Separated logic for creating new Consul client into its own file; .drone.yml: Deleted since no longer using drone
- Sign drone file
- Try to trigger build using protected mode

### Dockerfile
- Update go image to 1.11.4

### Docs
- Added Consul to demo
- Date command in demo is compatible with Mac and Linux
- increased interval in the example
- Edited Demonstration section
- Fixed a command in the demo section
- Added documentation on Demonstration; examples: Added a demo
- Added a brief intro and some badges; config/parse.go: Added more verbose error messages for rule parsing
- Rules dir env var now is correct
- Edited Nomad file
- Neglected to make html before committing
- Added links in Nomad section
- Fleshed out Usage section
- Added section in intro about architecture and statefulness
- Version bump

### Examples
- Minor modification to script

### Vendor
- now using go module


<a name="v0.0.22"></a>
## [v0.0.22] - 2018-12-11
### Chore
- go fmt
- More comments and examples
- Further commented exported assets

### Docs
- Added sidebar; README.md: Simplified and added link to docs page
- Changed build directory to be compatible with GH pages
- remove docs/build
- Added shadow to some images
- Increased documentation coverage
- Wrote the Installation section and some of the Intro
- Begin doc creation


<a name="v0.0.21"></a>
## [v0.0.21] - 2018-12-05
### Merge Requests
- Merge branch 'better-slack-formatting' into 'master'
- Merge branch 'master' into 'better-slack-formatting'


<a name="v0.0.20"></a>
## [v0.0.20] - 2018-12-05
### Chore
- Added more commenting to exported assets and formatting slack attachments better


<a name="v0.0.19"></a>
## [v0.0.19] - 2018-12-04

<a name="v0.0.18"></a>
## [v0.0.18] - 2018-12-04

<a name="v0.0.17"></a>
## [v0.0.17] - 2018-12-04
### Merge Requests
- Merge branch 'dynamic-body-field' into 'master'


<a name="v0.0.16"></a>
## [v0.0.16] - 2018-12-03

<a name="v0.0.15"></a>
## [v0.0.15] - 2018-12-03

<a name="v0.0.14"></a>
## [v0.0.14] - 2018-12-03
### Merge Requests
- Merge branch 'enable-rule-refresh' into 'master'


<a name="v0.0.13"></a>
## [v0.0.13] - 2018-12-03
### Chore
- Updated unit tests


<a name="v0.0.12"></a>
## [v0.0.12] - 2018-11-30
### Chore
- Include hits.hits array with state doc

### Merge Requests
- Merge branch 'improved-state-index' into 'master'


<a name="v0.0.11"></a>
## [v0.0.11] - 2018-11-30

<a name="v0.0.10"></a>
## [v0.0.10] - 2018-11-29

<a name="v0.0.9"></a>
## [v0.0.9] - 2018-11-29

<a name="v0.0.8"></a>
## [v0.0.8] - 2018-11-29

<a name="v0.0.7"></a>
## [v0.0.7] - 2018-11-28
### Chore
- formatted .go files

### Merge Requests
- Merge branch 'goreleaser' into 'master'


<a name="v0.0.6"></a>
## [v0.0.6] - 2018-11-28

<a name="v0.0.5"></a>
## [v0.0.5] - 2018-11-28
### Merge Requests
- Merge branch 'v0.0.1' into 'master'


<a name="v0.0.4"></a>
## [v0.0.4] - 2018-11-28

<a name="v0.0.3"></a>
## [v0.0.3] - 2018-11-28
### Chore
- Migrated dependencies to github

### Ci
- Build test-pr stage
- added some CI files

### Dockerfile
- Image built with Dockerfile runs the binary; scripts/build-binary.sh: Forgot to include version flags in cascade; scripts/docker-build.sh: Create a temporary Dockerfile for building only


<a name="v0.0.2"></a>
## v0.0.2 - 2018-11-27
### Bin
- Deleted

### Chore
- go fmt
- Delete changelog
- Added Changelog
- Added license
- minor logic fixes
- minor changes
- Fixed some mistakes

### Makefile
- Allow for relative paths when running git-chglog

### Vendor
- Updated deps
- Checked deps into vendoring


[Unreleased]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.35...HEAD
[v0.1.35]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.34...v0.1.35
[v0.1.34]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.33...v0.1.34
[v0.1.33]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.32...v0.1.33
[v0.1.32]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.31...v0.1.32
[v0.1.31]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.30...v0.1.31
[v0.1.30]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.29...v0.1.30
[v0.1.29]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.28...v0.1.29
[v0.1.28]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.27...v0.1.28
[v0.1.27]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.26...v0.1.27
[v0.1.26]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.25...v0.1.26
[v0.1.25]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.24...v0.1.25
[v0.1.24]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.23...v0.1.24
[v0.1.23]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.22...v0.1.23
[v0.1.22]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.21...v0.1.22
[v0.1.21]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.20...v0.1.21
[v0.1.20]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.19...v0.1.20
[v0.1.19]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.18...v0.1.19
[v0.1.18]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.17...v0.1.18
[v0.1.17]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.16...v0.1.17
[v0.1.16]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.15...v0.1.16
[v0.1.15]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.14...v0.1.15
[v0.1.14]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.13...v0.1.14
[v0.1.13]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.12...v0.1.13
[v0.1.12]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.11...v0.1.12
[v0.1.11]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.10...v0.1.11
[v0.1.10]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.9...v0.1.10
[v0.1.9]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.8...v0.1.9
[v0.1.8]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.7...v0.1.8
[v0.1.7]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.6...v0.1.7
[v0.1.6]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.5...v0.1.6
[v0.1.5]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.4...v0.1.5
[v0.1.4]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.2...v0.1.4
[v0.1.2]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.3...v0.1.2
[v0.1.3]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.1...v0.1.3
[v0.1.1]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.0...v0.1.1
[v0.1.0]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.22...v0.1.0
[v0.0.22]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.21...v0.0.22
[v0.0.21]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.20...v0.0.21
[v0.0.20]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.19...v0.0.20
[v0.0.19]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.18...v0.0.19
[v0.0.18]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.17...v0.0.18
[v0.0.17]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.16...v0.0.17
[v0.0.16]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.15...v0.0.16
[v0.0.15]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.14...v0.0.15
[v0.0.14]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.13...v0.0.14
[v0.0.13]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.12...v0.0.13
[v0.0.12]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.11...v0.0.12
[v0.0.11]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.10...v0.0.11
[v0.0.10]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.9...v0.0.10
[v0.0.9]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.8...v0.0.9
[v0.0.8]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.7...v0.0.8
[v0.0.7]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.6...v0.0.7
[v0.0.6]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.5...v0.0.6
[v0.0.5]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.4...v0.0.5
[v0.0.4]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.3...v0.0.4
[v0.0.3]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.0.2...v0.0.3

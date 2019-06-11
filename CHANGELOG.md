<a name="unreleased"></a>
## [Unreleased]


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
- Updated other files to reflect migration to go modules
- Using go modules instead of dep
- Using go modules instead of dep
- Updated license header to 2019; command/client.go: Separated logic for creating new Consul client into its own file; .drone.yml: Deleted since no longer using drone

### Ci
- Fix tag_filter to use glob [ci skip]
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


[Unreleased]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.6...HEAD
[v0.1.6]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.5...v0.1.6
[v0.1.5]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.4...v0.1.5
[v0.1.4]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.2...v0.1.4
[v0.1.2]: https://gitlab.morningconsult.com/mci/go-elasticsearch-alerts/compare/v0.1.1...v0.1.2
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

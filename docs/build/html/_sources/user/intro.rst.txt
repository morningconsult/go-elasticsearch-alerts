.. _introduction:

Introduction
============

Purpose
-------

Go Elasticsearch Alerts is a simple program that lets you generate
custom alerts on Elasticsearch data. It was built with simplicity
and flexibility in mind. While similar alternatives exist (e.g. 
`ElastAlert <https://github.com/Yelp/elastalert>`_), this program
has several distinct features:

- Greater query flexibility
- Multiple output methods (including Slack, email, and disk)
- Distributed operation via `Consul lock <https://www.consul.io/docs/commands/lock.html>`_
- Live rule updates
- Custom filters

License
-------
Copyright 2018 The Morning Consult, LLC or its affiliates. All Rights
Reserved.

Licensed under the Apache License, Version 2.0 (the "License"). You may
not use this file except in compliance with the License. A copy of the
License is located at

        https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
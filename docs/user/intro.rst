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

Overview
--------

Go Elasticsearch Alerts is made up of three main components:

- One or more *query handlers*;
- An *alert handler*; and
- One or more *alert outputs*.

At runtime, the process parses each :ref:`rule configuration file
<rule-configuration-file>` and starts a Goroutine - the *query handlers* -
for each rule. It starts another Goroutine - the *alert handler* - that
waits to receive new alerts from the query handlers. At intervals defined in
the rule, the query handler executes the query (also defined in the rule). If
Elasticsearch returns any data, it transforms the data based on the rule's
filters and sends the processed data to the alert handler. The alert handler
then sends the alerts to the specified *alert outputs* (e.g. Slack or email).
The query handlers will then pause until the next scheduled execution and
then repeat the process.

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
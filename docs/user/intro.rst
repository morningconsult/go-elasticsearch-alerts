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

This section summarizes the salient points of how Go Elasticsearch Alerts
was built and how it operates.

Architecture
~~~~~~~~~~~~

Go Elasticsearch Alerts is made up of three main components:

- One or more **query handlers**;
- An **alert handler**; and
- One or more **alert outputs**.

At runtime, the process parses :ref:`rule configuration files
<rule-configuration-file>` and starts a Goroutine for each rule. These
Goroutines are the **query handlers**. It starts another Goroutine - the
**alert handler** - that waits to receive new alerts from the query handlers.
At intervals defined in the rule, the query handler executes the Elasticsearch
query (also defined in the rule). If Elasticsearch returns any data, it
transforms the data based on the rule's filters and sends the processed data
to the alert handler. The alert handler then sends the alerts to the specified
**alert outputs** (e.g. Slack or email). The query handlers will then pause
until the next scheduled execution and then repeat the process.

.. _statefulness:

Statefulness
~~~~~~~~~~~~

Go Elasticsearch Alerts attempts to maintain the state of each query. This
ensures that if the process is restarted it will not immediately trigger the
query again; rather, it will trigger it when it was scheduled before the
process was killed. It achieves this by keeping records in a dedicated index
in your Elasticsearch host, henceforth referred to as the **state index**.
The documents stored in the state index represent a summary of the execution
of a query by a query handler. Each time the query handler triggers a query,
it writes such a document to the state index. An example is shown below.

.. _state-doc-example:

.. code-block:: json

  {
    "@timestamp": "2018-12-10T10:00:00Z",
    "next_query": "2018-12-10T10:30:00Z",
    "hostname": "ip-12-32-56-78",
    "rule_name": "example_errors",
    "hits_count": 0
  }

When the process is started, the query handler will attempt lookup the latest
document in the state index whose ``'rule_name'`` field matches the query
handler's rule name. If it finds a match, the query handler will schedule the
next execution of the query at the time given in the ``'next_query'`` field of
the matched document (e.g. at ``2018-12-10T10:30:00Z`` in the :ref:`example 
above <state-doc-example>`). If value of ``'next_query'`` is in the past, it
will execute the query immediately.

Immediately following the execution of a query, the query handler will write a
new document to the state index where the value of the ``'next_query'`` field
will be equal to the next time that the query should be executed per the
schedule defined in the rule. Additionally, it will include the number of hits
Elasticsearch returned in the response to the query and the actual hits
themselves.

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
.. _setup:

.. role:: red

.. role:: code-no-background

Setup
=====

Before reading further, make sure that Go Elasticsearch Alerts is
:ref:`installed <install>`.

Configuration Files
-------------------

This program requires some configuration files: a `main configuration file`_
and one or more `rule configuration files <#rule-configuration-file>`__.

Main Configuration File
-----------------------

The main configuration file is a JSON file which configures some of the
general behavior of the program. Specifically, it is used to configure
communication with Elasticsearch and distributed operation. The program
will not operate without this file.

The program will look for the main configuration file at
``/etc/go-elasticsearch-alerts/config.json`` by default. If you wish to keep
this file elsewhere, you can specify its location with the
``GO_ELASTICSEARCH_ALERTS_CONFIG_FILE`` environment variable.

Example
~~~~~~~

.. code-block:: json

  {
    "elasticsearch": {
      "server": {
        "url": "https://my.elasticsearch.com"
      },
      "client": {
        "tls_enabled": true,
        "ca_cert": "/tmp/cacert.pem",
        "client_cert": "/tmp/client_cert.pem",
        "client_key": "/tmp/client_key.pem"
      }
    },
    "distributed": true,
    "consul": {
      "consul_lock_key": "go-elasticsearch-alerts/leader",
      "consul_http_addr": "http://127.0.0.1:8500"
    }
  }

Main File Parameters
~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`elasticsearch` (`Elasticsearch
  <#elasticsearch-parameters>`__: ``<nil>``) - Configures how the program
  communcates with Elasticsearch. See the
  `Elasticsearch <#elasticsearch-parameters>`__ section for more details.
  This field is required.
- :code-no-background:`distributed` (bool: ``false``) - Whether this
  application will be run in a distributed fashion. If this is set to
  ``true`` then the ``consul`` field will also be required since this
  application uses the `lock feature
  <https://www.consul.io/docs/commands/lock.html>`__ of HashiCorp's `Consul
  <https://www.consul.io/>`__ service for synchronization between nodes.
  This field is optional. For more information on distributed operation,
  see the :ref:`distributed usage <distributed>` section.
- :code-no-background:`consul` (`Consul <#consul-parameters>`__: ``<nil>``)
  - Configures the Consul client. The program will use this client to
  communicate with your Consul server for synchronization between nodes. This
  field is required if ``distributed`` is ``true``.

``elasticsearch`` Parameters
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`server` (`Server <#server-parameters>`__: ``<nil>``)
  - Specifies information pertaining to your Elasticsearch server. See the
  `Server <#server-parameters>`__ section for more information. This field
  is always required.
- :code-no-background:`client` (`Client <#client-parameters>`__: ``<nil>``)
  - Configures the HTTP client with which the program will communicate with
  Elasticsearch. See the `Client <#client-parameters>`__ section for more
  information. This field is always required.

``consul`` Parameters
~~~~~~~~~~~~~~~~~~~~~

Note: All of these values should be strings. For example, even if the value
is technically a Boolean field such as ``true``, you should provide a string
(e.g. ``"true"``).

- :code-no-background:`consul_lock_key` (string: ``""``) - The name of the key to be assigned
  to the Consul lock. This field is always required.
- ``consul_http_addr`` (string: ``""``) - The URL of your Consul server.
  This field is always required.
- :code-no-background:`consul_http_token` (string: ``""``) - The API access token required
  when access control lists (ACLs) are enabled. This field is
  optional.\ :red:`*`
- :code-no-background:`consul_http_ssl` (string: ``"false"``) - A boolean
  value (default is false) that enables the HTTPS URI scheme and SSL
  connections to the HTTP API. This field is optional.\ :red:`*`
- :code-no-background:`consul_http_ssl_verify` (string: ``""``) - A boolean
  value (default true) to specify SSL certificate verification; setting this
  value to false is not recommended for production use. This field is
  optional.\ :red:`*`
- :code-no-background:`consul_cacert` (string: ``""``) - Path to a CA file to
  use for TLS when communicating with Consul. This field is
  optional.\ :red:`*`
- :code-no-background:`consul_capath` (string: ``""``) - Path to a directory
  of CA certificates to use for TLS when communicating with Consul. This
  field is optional.\ :red:`*`
- :code-no-background:`consul_client_cert` (string: ``""``) - Path to a client
  cert file to use for TLS when verify_incoming is enabled. This field is
  optional.\ :red:`*`
- :code-no-background:`consul_client_key` (string: ``""``) - Path to a client
  key file to use for TLS when verify_incoming is enabled. This field is
  optional.\ :red:`*`
- :code-no-background:`consul_tls_server_name` (string: ``""``) - The server
  name to use as the SNI host when connecting via TLS. This field is
  optional.\ :red:`*`

:red:`*`\ This field can be specified using its corresponding `environment
variable <https://www.consul.io/docs/commands/index.html#environment-variables>`__
instead. The environment variable takes precedence.

``server`` Parameters
~~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`url` (string: ``""``) - The URL of your Elasticsearch
  instance. This field is always required.

``client`` Parameters
~~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`tls_enabled` (bool: ``false``) - Whether the application
  should use TLS when communicating with your Elasticsearch server. This field
  is optional.
- :code-no-background:`ca_cert` (string: ``""``) - Path to a PEM-encoded CA
  certificate file on the local disk. This file is used to verify the
  Elasticsearch server's SSL certificate.
- :code-no-background:`client_cert` (string: ``""``) - Path to a PEM-encoded
  client certificate on the local disk. This file is used for TLS
  communication with the Elasticsearch server.
- :code-no-background:`client_key` (string: ``""``) - Path to an unencrypted,
  PEM-encoded private key on disk which corresponds to the matching client
  certificate.
- :code-no-background:`server_name` (string: ``""``) - Name to use as the SNI
  host when connecting via TLS.

Rule Configuration File
-----------------------

The rule configuration files are JSON files which define your alerts. The
program will look for the rule configuration files in the
``/etc/go-elasticsearch-alerts/rules`` directory by default. If you wish to
keep these files in a different directory, you can specify this directory
with the ``GO_ELASTICSEARCH_ALERTS_CONFIG_FILE`` environment variable. All of
these files should be valid JSON and their file names should have a ``.json``
extension. There must be at least one rule for the program to operate.

.. _rule-example:

Example
~~~~~~~

.. code-block:: json

  {
    "name": "Filebeat Errors",
    "index": "filebeat-*",
    "schedule": "@every 10m",
    "body": {
      "query": {
        "bool": {
          "must": [
            { "query_string" : {
              "query" : "*",
              "fields" : [ "system.syslog.message", "message" ]
            } }
          ]
        }
      },
      "aggs": {
        "hostname": {
          "terms": {
            "field": "system.syslog.hostname",
            "min_doc_count": 1
          }
        }
      },
      "size": 20,
      "_source": "system.syslog"
    },
    "body_field": "hits.hits._source",
    "filters": [
      "aggregations.service_name.buckets",
      "aggregations.service_name.buckets.program.buckets"
    ],
    "outputs": [
      {
        "type": "slack",
        "config" : {
          "webhook": "https://hooks.slack.com/ASDFASDF",
          "text": "New errors",
        }
      },
      {
        "type": "email",
        "config": {
          "to": [
            "you@example.com"
          ]
          "from": "me@example.com",
          "host": "smtp.gmail.com",
          "port": 587
        }
      }
    ]
  }

In the example above, the application would execute the following query
(illustrated as a ``cURL`` request below) to Elasticsearch every ten minutes,
group by ``hits.hits._source``, ``aggregations.service_name.buckets``, and
``aggregations.service_name.buckets.program.buckets``, and write the results
to Slack and local disk.

.. _curl-request:

.. code-block:: shell

  $ curl http://https://my.elasticsearch.com/filebeat-*/_search \
    --header "Content-Type: application/json" \
    --data '{
      "query": {
        "bool": {
          "must": [
            { "query_string" : {
              "query" : "*",
              "fields" : [ "system.syslog.message", "message" ]
            } }
          ]
        }
      },
      "aggs": {
        "hostname": {
          "terms": {
            "field": "system.syslog.hostname",
            "min_doc_count": 1
          }
        }
      },
      "size": 20,
      "_source": "system.syslog"
    }'

Rule File Parameters
~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`name` (string: ``""``) - The name of the rule (e.g.
  ``"Filebeat Errors"``). This field is required.
- :code-no-background:`index` (string: ``""``) - The index to be queried.
  This field is required.
- :code-no-background:`schedule` (string: ``""``) - When the query should be
  executed. This should be a `cron <https://en.wikipedia.org/wiki/Cron>`__
  string. This program uses `github.com/robfig/cron
  <https://godoc.org/github.com/robfig/cron>`__ to parse the cron schedule,
  so please refer to it for specifics on how to write a proper cron schedule.
- :code-no-background:`body` (JSON object: ``<nil>``) - The body of the
  `search query
  <https://www.elastic.co/guide/en/elasticsearch/reference/current/search-request-body.html>`__
  request. When the job is triggered, the program will pass this exact JSON as
  data in the request to the ``<index>/_search`` endpoint. The value of this
  field will dictate the structure of the Elasticsearch response data and
  therefore will dictate the set of potential values for the ``filters`` and
  ``body_field`` sections. It is recommendeded that you manually run this
  query (for an example, see the :ref:`cURL request <curl-request>` above)
  and understand the structure of the response data before setting the
  ``filters`` and ``body_field`` sections.
- :code-no-background:`filters` ([]string: ``[]``) - How the response to this
  query should be grouped. How the group data will be presented depends on
  the output method(s) used. More information on this field is provided in the
  `filters`_ section. More information on how the response data appears in the
  different output media is provided in the :ref:`Alerting <alerting>` section.
- :code-no-background:`body_field` (string: ``"hits.hits._source"``) - The
  field on which to group the response. The elements of the response data
  that match the value of this field will be stringified and concatenated
  before being sent to the provided output(s). This field is optional. If
  not specified, the program will group by the field ``hits.hits._source``
  by default. See the :ref:`Alerting <alerting>` section for more information
  on this field.
- :code-no-background:`outputs` ([]\ `Output <#outputs-parameters>`__: ``[]``)
  - The media by which alerts should be sent. See the `Output
  <#outputs-parameters>`__ section for more details. At least one output must
  be specified.

``outputs`` Parameters
~~~~~~~~~~~~~~~~~~~~~~

The :code-no-background:`outputs` parameter of the rule file specifies where
the results of the queries should be sent. Each rule should have at least one
output. Currently, three output types are supported:
`Slack <#slack-output-parameters>`__, `email <#email-output-parameters>`__,
and `file <#file-output-parameters>`__. The exact specifications of this field
will depend on the output type.

- :code-no-background:`type` (string: ``""``) - The type of output. Currently,
  only ``slack``, ``file``, and ``email`` are supported. This field is always
  required.
- :code-no-background:`config` (JSON object: ``<nil>``) - Configurations
  specific to the output type. This field is alwyas required.

Slack Output Parameters
~~~~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`webhook` (string: ``""``) - The Slack webhook where
  error alerts will be sent. This field is required.
- :code-no-background:`text` (string: ``""``) - Text to be sent with the
  Slack message.

Email Output Parameters
~~~~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`host` (string: ``""``) - The SMTP server host (e.g.
  ``smtp.gmail.com``). This field is required.
- :code-no-background:`port` (int: ``0``) - The SMTP server port (e.g. ``587``
  for Gmail). This field is required.
- :code-no-background:`from` (string: ``""``) - The "from" email address. This
  field is required.
- :code-no-background:`to` ([]string: ``[]``) - The "to" addresses to which
  email alerts will be sent. At least one address is required.
- :code-no-background:`username` (string: ``""``) - The username with which
  the SMTP client will authenticate to the host. If you do not wish to specify
  the username in the configuration file, you can set the password using the
  ``GO_ELASTICSEARCH_ALERTS_SMTP_USERNAME`` environment variable. This field
  is required (either in the configuration file or in the environment
  variable).
- :code-no-background:`password` (string: ``""``) - The password with which the
  SMTP client will authenticate to the host. If you do not wish to specify the
  password in the configuration file, you can set the password using the
  ``GO_ELASTICSEARCH_ALERTS_SMTP_PASSWORD`` environment variable. This field is
  required (either in the configuration file or in the environment variable).

File Output Parameters
~~~~~~~~~~~~~~~~~~~~~~

- :code-no-background:`file` (string: ``""``) - The file to which alerts will
  be written. This field is required.

Filters
-------

Filters are used to group the data that Elasticsearch responds to a query with.
They are used to provide a brief summary of the response data.

For example, given the :ref:`example <rule-example>` above, assume that after
the job triggers and executes the query Elasticsearch responds with the
following data:

.. code-block:: json

  {
    "hits": {
      "hits": [
        {
          "_source": {
            "@timestamp": "2018-12-10 11:00:00",
            "hello": "world"
          }
        },
        {
          "_source": {
            "@timestamp": "2018-12-10 11:01:00",
            "foo": "bar"
          }
        }
      ]
    },
    "aggregations": {
      "service_name": {
        "buckets": [
          {
            "key": "nomad",
            "doc_count": 10,
            "program": {
              "buckets": [
                {
                  "key": "app-1",
                  "doc_count": 4
                },
                {
                  "key": "app-2",
                  "doc_count": 6
                }
              ]
            }
          },
          {
            "key": "consul",
            "doc_count": 4,
            "program": {
              "buckets": [
                {
                  "key": "node-1",
                  "doc_count": 3
                },
                {
                  "key": "node-2",
                  "doc_count": 1
                }
              ]
            }
          }
        ]
      }
    }
  }

Upon receiving this JSON data, the program will group the data by
``"aggregations.service_name.buckets"`` and
``"aggregations.service_name.buckets.program.buckets"`` and send the results
via Slack and email as shown below.

Slack Output Example
~~~~~~~~~~~~~~~~~~~~

.. image:: ../images/slack.png

Email Output Example
~~~~~~~~~~~~~~~~~~~~

.. image:: ../images/email.png

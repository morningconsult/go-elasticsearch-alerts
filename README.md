# go-elasticsearch-alerts

Elasticsearch Alerting Daemon

## Installation
```shell
$ go get gitlab.morningconsult.com/mci/go-elasticsearch-alerts
```

# Setup

This application requires several configuration files: a [main configuration file](#main-configuration-file) and one or more [rule configuration files](#rule-configuration-files) The main configuration file should be placed in a directory (default: `/etc/go-elasticsearch-alerts/config.json`) with 

## Main Configuration File

The main configuration file is used to specify information regarding your ElasticSearch instance how this application will interact with it. The application will look for this file at `/etc/go-elasticsearch-alerts/config.json` by default, but if you wish to keep it elsewhere you can specify the location of this file using the `GO_ELASTICSEARCH_ALERTS_CONFIG_FILE` environment variable.

### Example

This example shows a sample main configuration file.

```json
{
  "elasticsearch": {
    "server": {
      "url": "https://my.elasticsearch.com",
      "state_index": "go_elasticsearch_status"
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
    "consul_http_address": "http://127.0.0.1:8500",
    "consul_http_ssl": true,
    "consul_cacert": "/tmp/cacert_consul.pem",
    "consul_client_cert": "/tmp/client_cert_consul.pem",
    "consul_client_key": "/tmp/client_key_consul.pem"
  }
}
```
* `elasticsearch` ([ElasticSearch](#elasticsearch-parameters): `<nil>`) - Configures the ElasticSearch client and specifies server parameters. See the [ElasticSearch](#elasticsearch-parameters) section for more details. This field is required.
* `distributed` (bool: `false`) - Whether this application should will be distributed across multiple processes. If this is set to `true`, the `consul` field is also required since this application uses the [Consul lock](https://www.consul.io/docs/commands/lock.html) for synchronization. This field is optional.
* `consul` ([Consul](#consul-parameters): `<nil>`) - Configures the Consul client if this application is distributed. This field is only required when `distributed` is set to `true`.

### `elasticsearch` parameters
* `server` ([Server](#server-parameters): `<nil>`) - Specifies ElasticSearch server information. See the (Server)[#server-parameters] section for more information. This field is always required.
* `client` ([Client](#client-parameters): `<nil>`) - Configures the HTTP client with which the process will communicate with Elasticsearch. See the (Client)[#client-parameters] section for more informiation. This field is always required.

### `consul` parameters
* `consul_lock_key` (string: `""`) - The name of the key to be assigned to the Consul lock. This field is always required.
* `consul_http_address` (string: `""`) - The URL of your Consul server. This field is always required.
* `consul_http_token` (string: `""`) - The API access token required when access control lists (ACLs) are enabled. This field is optional.**\***
* `consul_http_ssl` (bool: `false`) - A boolean value (default is false) that enables the HTTPS URI scheme and SSL connections to the HTTP API. This field is optional.**\***
* `consul_http_ssl_verify` (string: `""`) - A boolean value (default true) to specify SSL certificate verification; setting this value to false is not recommended for production use. This field is optional.**\***
* `consul_cacert` (string: `""`) - Path to a CA file to use for TLS when communicating with Consul. This field is optional.**\***
* `consul_capath` (string: `""`) - Path to a directory of CA certificates to use for TLS when communicating with Consul. This field is optional.**\***
* `consul_client_cert` (string: `""`) - Path to a client cert file to use for TLS when verify_incoming is enabled. This field is optional.**\***
* `consul_client_key` (string: `""`) - Path to a client key file to use for TLS when verify_incoming is enabled. This field is optional.**\***
* `consul_tls_server_name` (string: `""`) - The server name to use as the SNI host when connecting via TLS. This field is optional.**\***

**\*** This field can be specified using its corresponding [environment variable](https://www.consul.io/docs/commands/index.html#environment-variables) instead. If this field and its corresponding environment variable are both set, the environment variable takes precedence.

### `server` parameters

* `url` (string: `""`) - The URL of your ElasticSearch instance. This field is always required.
* `state_index` (string: `"go_elasticsearch_alerts_state"`) - The ElasticSearch index where the records of when each rule is schedule to run next is stored. If this index does not already exist, the application will attempt to create it at runtime. This is used to ensure that if the process is killed at any point it can lookup the next scheduled run when started again. This field is optional. If not specified, the default will be `go_elasticsearch_alerts_state`.

### `client` parameters
* `tls_enabled` (bool: `false`) - Whether the application should use TLS when communicating with your ElasticSearch instance. This field is optional.
* `ca_cert` (string: `""`) - Path to a PEM-encoded CA certificate file on the local disk. This file is used to verify the ElasticSearch server's SSL certificate.
* `client_cert` (string: `""`) - Path to a PEM-encoded client certificate on the local disk. This file is used for TLS communication with the ElasticSearch server.
* `client_key` (string: `""`) - Path to an unencrypted, PEM-encoded private key on disk which corresponds to the matching client certificate.
* `server_name` (string: `""`) - Name to use as the SNI host when connecting via TLS.

### `distributed` parameters
* `consul_address` (string: `""`) - The URL of your Consul server.
* `consul_lock_key` (string: `""`) - They name of the key that Consul should use to create the lock.

### Rule Configuration Files

The rule configuration files are used to configure what ElasticSearch queries will be run, how often they will be run, how the data will be transformed, and how the transformed data will be output. These files should be JSON format. The application will look for the rule files at `/etc/go-elasticsearch-alerts/rules` by default, but if you wish to keep them elsewhere you can specify this directory using the `GO_ELASTICSEARCH_ALERTS_RULES_DIR` environment variable.

### Example

```json
{
  "name": "filebeat_errors",
  "index": "filebeat-*",
  "schedule": "* */10 * * * *",
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
  "filters": [
    "aggregations.service_name.buckets",
    "aggregations.service_name.buckets.program.buckets"
  ],
  "outputs": [
    {
      "type": "slack",
      "config" : {
        "webhook": "https://slack.webhooks.foo/asdf",
        "channel": "#error-alerts",
        "text": "New errors",
        "emoji": ":hankey:"
      }
    },
    {
      "type": "file",
      "config": {
        "file": "/tmp/errors.log"
      }
    }
  ]
}
```

In the example above, the application would execute the following query (show as a `cURL` request) to ElasticSearch every ten minutes, group by `aggregations.service_name.buckets` and `aggregations.service_name.buckets.program.buckets`, and send the results to a Slack and write them to local disk.

```shell
$ curl http://<your_elasticsearch_host>/filebeat-*/_search \
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
```

### Rule configuration file parameters

* `name` (string: `""`) - The name of the rule. This should be unique and have no spaces. This field is required.
* `index` (string: `""`) - The index to be queried. This field is required.
* `schedule` (string: `""`) - The schedule of when the query will be executed in [cron syntax](https://en.wikipedia.org/wiki/Cron).
* `body` (JSON object: `<nil>`) - The body of the [search query](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-request-body.html) request. This should be exactly what you would include in an ElasticSearch `_search` request to the index specified above. This value will dictate the layout of the data that your ElasticSearch instance sends to this application; therefore, the subsequent `filters` section is dictated by this section. It is recommended that you manually run this query and understand the structure of the response data before writing the `filters` section.
* `filters` (\[\]string: `<nil>`) - How the response to this query should be grouped. More information on this field is provided in the [filters](#filters) section. This field is optional. If no filters are provided, only elements of the `hits.hits._source` field of the response will be recorded.
* `outputs` (\[\][Output](#outputs-parameter): `<nil>`) - Specifies the outputs to which the results of the query should be written. See the [Output](#output-parameter) section for more details. At least one output must be specified.

### Filters

The application will group the response to the ElasticSearch query by each element of the `filters` field and include each result of the filters as a separate record. For example, given the [rule file above](#example) let's assume that ElasticSearch returns the following in response to the query:
```json
{
  "hits": {
    "hits": [
      {
        "_source": {
          "some": {
            "important": "information"
          }
        }
      },
      {
        "_source": {
          "more": {
            "important": "info!"
          }
        }
      }
    ]
  },
  "aggregations": {
    "service_name": {
      "buckets": [
        {
          "key": "foo",
          "doc_count": 10,
          "program": {
            "buckets": [
              {
                "key": "bim",
                "doc_count": 4
              },
              {
                "key": "baz",
                "doc_count": 6
              }
            ]
          }
        },
        {
          "key": "bar",
          "doc_count": 4,
          "program": {
            "buckets": [
              {
                "key": "ayy",
                "doc_count": 2
              },
              {
                "key": "lmao",
                "doc_count": 2
              }
            ]
          }
        }
      ]
    }
  }
}
```

Also given the filters `aggregations.service_name.buckets` and `aggregations.service_name.buckets.program.buckets` and that the Slack output method is used, the application will make the following request to Slack (shown as a `cURL` request) after running the query and receiving the aforementioned data (note that if the response from ElasticSearch includes the array field `hits.hits`, the `_source` field of each element will be included in the output by default, regardless of the `filters`):

```shell
$ curl https://slack.webhooks.foo/asdf \
  --request POST \
  --header "Content-Type: application/json" \
  --data '{
  "channel": "#error-alerts",
  "text": "New errors",
  "emoji": ":hankey:",
  "attachments": [
    {
      "fallback": "filebeat_errors",
      "pretext": "aggregations.service_name.buckets",
      "fields": [
        {
          "title": "foo",
          "value": 10,
          "short": true
        },
        {
          "title": "bar",
          "value": 4,
          "short": true
        }
      ]
    },
    {
      "fallback": "filebeat_errors",
      "pretext": "aggregations.service_name.buckets.program.buckets",
      "fields": [
        {
          "title": "foo - bim",
          "value": 4,
          "short": true
        },
        {
          "title: "foo - baz",
          "value": 6,
          "short": true
        },
        {
          "title": "bar - ayy",
          "value": 2,
          "short": true
        },
        {
          "title": "bar - lmao",
          "value": 2,
          "short": true
        }
      ]
    },
    {
      "fallback": "hits.hits._source",
      "text": "{
        \"some\": {
          \"important\": \"information\"
        }
      }
      ----------------------------------------
      {
        \"more\": {
          \"important\": \"info!\"
        }
      }
      "
    }
  ]
}'
``` 

Note: The last element of the `filter` value should be an array with both the `key` and `doc_count` fields (e.g. if you use `aggregations.hostname.buckets`, then `buckets` should be an array).

### `outputs` parameter

The `outputs` parameter of the rule file specifies where the results of the queries should be written. Each rule file should have at least one output. Currently, two outputs are supported: [Slack](#slack-output-configuration-parameters) and [File](#file-output-configuration-parameters). The exact structure of this field will depend on the output type.

* `type` (string: `""`) - The type of output. Currently, only `slack` and `file` are supported. This field is always required.
* `config` (JSON object: `<nil>`) - The configuration parameters of the output type. The parameters required in this section are specific to the output type. This field is always required.

#### Slack Output Configuration Parameters

* `webhook` (string: `""`) - The Slack webhook where error alerts will be sent. This field is required.
* `channel` (string: `""`) - The Slack channel where error alerts will be posted. This field is optional.
* `username` (string: `""`) - The Slack bot username which will be used to post new error alerts. This field is optional.
* `text` (string: `""`) - Text that will be included in the Slack posts. This field is optional.
* `emoji` (string: `""`) - The emoji that will be included in the Slack posts. This field is optional.

#### File Output Configuration Parameters

* `file` (string: `""`) - The file to which alerts should be written. This field is required.

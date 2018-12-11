# go-elasticsearch-alerts

[![Build Status](https://ci.morningconsultintelligence.com/api/v1/teams/oss/pipelines/go-elasticsearch-alerts/jobs/build-release/badge)](https://ci.morningconsultintelligence.com/teams/oss/pipelines/go-elasticsearch-alerts) [![Go Documentation](https://img.shields.io/badge/godoc-reference-blue.svg)][godocs] [![Go Report Card](https://goreportcard.com/badge/github.com/morningconsult/go-elasticsearch-alerts)](https://goreportcard.com/report/github.com/morningconsult/go-elasticsearch-alerts)

[godocs]: https://godoc.org/github.com/morningconsult/go-elasticsearch-alerts

A daemon for generating alerts on Elasticsearch data in real-time.

Further details on setup and usage can be found in the [project documentation](https://morningconsult.github.io/go-elasticsearch-alerts).

## Installation

### Manually

You can download your preferred variant of the binary from the [releases page](https://github.com/morningconsult/go-elasticsearch-alerts/releases).

### Using `go get`

You can build the binary via `go get` with 

```shell
$ go get github.com/morningconsult/go-elasticsearch-alerts
```

### Using Docker

If you do not have Go installed locally, you can still build the binary if you have Docker installed. Simply clone this repository and run `make docker` to build the binary within a Docker container and output it to the local directory.

You can cross-compile the binary using the `TARGET_GOOS` and `TARGET_GOARCH` environment variables. For example, if you wish to compile the binary for a 64-bit (x86-64) Windows machine, run the following command:

```shell
$ TARGET_GOOS="windows" TARGET_GOARCH="amd64" make docker
```

The binary will be output to `bin` in the local directory.

# Setup

This application requires several configuration files: a [main configuration file](#main-configuration-file) and one or more [rule configuration files](#rule-configuration-files). The main configuration file is used to configure general behavior of the application. The rule files are used to define your alerts (e.g. what queries are executed, when they are executed, where the results shall be sent, etc.).

## Main Configuration File

The main configuration file is used to specify:
* Information pertaining to your Elasticsearch instance;
* How the application will interact with your Elasticsearch instance;
* Whether it is to be run in a distributed fashion; and
* If distributed, how the application will communicate with your Consul instance (used for synchronization).

The application will look for this file at `/etc/go-elasticsearch-alerts/config.json` by default, but if you wish to keep it elsewhere you can specify the location of this file using the `GO_ELASTICSEARCH_ALERTS_CONFIG_FILE` environment variable.

### Example

This example shows a sample main configuration file.

```json
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
    "consul_http_addr": "http://127.0.0.1:8500",
    "consul_http_ssl": "true",
    "consul_cacert": "/tmp/cacert_consul.pem",
    "consul_client_cert": "/tmp/client_cert_consul.pem",
    "consul_client_key": "/tmp/client_key_consul.pem"
  }
}
```

### Rule Configuration Files

The rule configuration files are used to configure what Elasticsearch queries will be run, how often they will be run, how the data will be transformed, and how the transformed data will be output. These files should be JSON format. The application will look for the rule files at `/etc/go-elasticsearch-alerts/rules` by default, but if you wish to keep them elsewhere you can specify this directory using the `GO_ELASTICSEARCH_ALERTS_RULES_DIR` environment variable.

### Example

```json
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

In the example above, the application would execute the following query (illustrated by the `cURL` request below) to Elasticsearch every ten minutes, group by `aggregations.service_name.buckets` and `aggregations.service_name.buckets.program.buckets`, and write the results to Slack and local disk.

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

## Usage

Once your configuration files have been setup, to run the program
simply execute the binary

```shell
$ ./go-elasticsearch-alerts
```

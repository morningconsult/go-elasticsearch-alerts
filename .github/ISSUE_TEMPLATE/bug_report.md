---
name: Bug Report
about: You're experiencing an issue with this project that is different than the documented behavior.

---

When filing a bug, please include the following headings if possible. Any example text in this template can be deleted.

#### Overview of the Issue

A paragraph or two about the issue you're experiencing.

#### Reproduction Steps

Steps to reproduce this issue, eg:

1. Run `TARGET_GOOS=windows TARGET_GOARCH=amd64 make docker`
1. Move the binary to a location on your path
1. Run `go-elasticsearch-alerts`
1. View error

### A copy of your main configuration file and your rule configuration file(s)

**Main Configuration File**
```json
{
  "elasticsearch": {
    "server": {
      "url": "http://127.0.0.1:9200"
    },
    "client": {
      "tls_enabled": false
    }
  },
  "distributed": true,
  "consul": {
    "consul_http_addr": "http://127.0.0.1:8500",
    "consul_lock_key": "go-elasticsearch-alerts/lock"
  }
}
```

**Rule Configuration File**
```json
{
  "name": "Example Error",
  "index": "test-index",
  "schedule": "@every 2m",
  "body": {
    "query": {
      "bool": {
        "must": [
          { "term" : { "source" : { "value" : "/var/log/system.log" } } }
        ],
        "filter": [
          { "range" : { "@timestamp" : { "gte" : "now-2m/m" } } }
        ]
      }
    },
    "aggregations": {
      "hostname": {
        "terms": {
          "field": "system.syslog.hostname",
          "min_doc_count": 1
        }
      }
    },
    "size": 20,
    "sort": [
      { "@timestamp": "desc" }
    ]
  },
  "filters": [
    "aggregations.hostname.buckets"
  ],
  "outputs": [
    {
      "type": "file",
      "config": {
        "file": "/dev/stdout"
      }
    }
  ]
}
```

### Operating system and Environment details

OS, Architecture, and any other information you can provide about the environment.

### Log Fragments

Include appropriate log fragments. 
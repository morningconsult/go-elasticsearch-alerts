#!/bin/bash

set -e

INDEX="test-index"
ES_URL="http://127.0.0.1:9200"
CONSUL_URL="http://127.0.0.1:8500"

## Start Elasticsearch container
docker-compose up -d elasticsearch-gea

echo "==> Starting Elasticsearch health checks."

## Wait until Elasticsearch is healthy by checking for health 10 times
for i in {0..10}
do
  if [ $i -gt 9 ]
  then
    echo "==> Elasticsearch is not healthy after 10 attempts"
    exit 1
  fi

  echo "==> Performing Elasticsearch health check..."

  STATUS=$( curl -s "${ES_URL}/_cluster/health" | jq .status )
  if [ "${STATUS}" == '"green"' ]
  then
    break
  fi

  echo "==> Elasticsearch health check failed. Retrying in 20 seconds."

  sleep 20

done

echo "==> Elasticsearch is healthy. Creating index \"${INDEX}\"..."

## Create the test index
curl "${ES_URL}/${INDEX}" \
  -s \
  -H "Content-Type: application/json" \
  -X PUT \
  -d '{
    "settings": {
      "number_of_shards": 1
    },
    "mappings": {
      "_doc": {
        "properties": {
          "@timestamp": { "type": "date" },
          "source": { "type": "keyword" },
          "system": { 
            "properties": {
              "syslog": {
                "properties": {
                  "hostname": { "type" : "keyword" },
                  "message": { "type" : "keyword" }
                }
              }
            }
          }
        }
      }
    }
  }' > /dev/null


NOW="$( date +%s )000"
cat <<EOF > /tmp/gea-payload-1.json
{
  "@timestamp": "${NOW}",
  "source": "/var/log/system.log",
  "system": {
    "syslog": {
      "hostname": "ip-127-0-0-1",
      "message": "You got an error buddy!"
    }
  }
}
EOF

cat <<EOF > /tmp/gea-payload-2.json
{
  "@timestamp": "${NOW}",
  "source": "/var/log/errors.log",
  "system": {
    "syslog": {
      "hostname": "ip-172-32-0-1",
      "message": "Another error!"
    }
  }
}
EOF

echo "==> Writing some test data to Elasticsearch..."

for f in /tmp/gea-payload-*.json; do
  # Write a document to the new index
  curl "${ES_URL}/${INDEX}/_doc" \
    -s \
    -H "Content-Type: application/json" \
    -X POST \
    -d "@${f}" > /dev/null
done

## Clean up test data files
rm /tmp/gea-payload-*.json

## Start up Consul

echo "==> Done writing Elasticsearch data. Starting Consul..."

docker-compose up -d consul-gea

sleep 2

## Wait until Consul is healthy by checking for health 10 times
for i in {0..10}
do
  if [ $i -gt 9 ]
  then
    echo "Consul is not healthy after 10 attempts"
    exit 1
  fi

  STATUS=$( curl -s "${CONSUL_URL}/v1/status/leader" )
  if [ "${STATUS}" == '"127.0.0.1:8300"' ]
  then
    break
  fi

  sleep 2
done

sleep 2

echo "==> Done writing Elasticsearch data. Starting Go Elasticsearch Alerts container..."

docker-compose up go-elasticsearch-alerts

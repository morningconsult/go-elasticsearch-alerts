#!/bin/bash

set -e

INDEX="test-index"

## Start Elasticsearch container
docker-compose up -d elasticsearch_gea

echo "==> Waiting for Elasticsearch to be healthy..."

## Wait until Elasticsearch is healthy by checking for health 10 times
for i in {0..10}
do
  if [ $i -gt 9 ]
  then
    echo "Elasticsearch is not healthy after 10 attempts"
    exit 1
  fi

  # CURL=$( curl http://127.0.0.1:9200/_cluster/health )
  # echo $CURL
  STATUS=$( curl -s http://127.0.0.1:9200/_cluster/health | jq .status )
  if [ "${STATUS}" == '"green"' ]
  then
    break
  fi

  sleep 3
done

echo "==> Elasticsearch is healthy. Creating index \"${INDEX}\"..."

## Create the test index
curl "http://127.0.0.1:9200/${INDEX}" \
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
  curl "http://127.0.0.1:9200/${INDEX}/_doc" \
    -s \
    -H "Content-Type: application/json" \
    -X POST \
    -d "@${f}" > /dev/null
done

## Clean up test data files
rm /tmp/gea-payload-*.json

sleep 2

echo "==> Done writing Elasticsearch data. Starting Go Elasticsearch Alerts container..."

docker-compose up go-elasticsearch-alerts

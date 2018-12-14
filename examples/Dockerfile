FROM golang:1.11.3-alpine3.8

RUN set -e; \
  apk add -qU --no-cache git; \
  go get github.com/morningconsult/go-elasticsearch-alerts;

COPY . .

CMD ["go-elasticsearch-alerts"]

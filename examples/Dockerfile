FROM golang:1.11.4-alpine3.8

RUN set -e; \
  apk add -qU --no-cache git bzr make;

WORKDIR /build

RUN set -e; \
  git clone https://github.com/morningconsult/go-elasticsearch-alerts; \
  cd go-elasticsearch-alerts; \
  GO111MODULE=on CGO_ENABLED=0 make; \
  mv ./bin/go-elasticsearch-alerts /usr/local/bin;

COPY . .

CMD ["go-elasticsearch-alerts"]

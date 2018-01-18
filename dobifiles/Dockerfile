ARG     GOLANG_VERSION
FROM    golang:${GOLANG_VERSION:-1.9.2-alpine}

RUN     apk add -U curl git bash

ARG     FILEWATCHER_SHA=2e12ea42f6c8c089b19e992145bb94e8adaecedb
RUN     go get -d github.com/dnephin/filewatcher && \
        cd /go/src/github.com/dnephin/filewatcher && \
        git checkout -q "$FILEWATCHER_SHA" && \
        go build -v -o /usr/bin/filewatcher . && \
        rm -rf /go/src/* /go/pkg/* /go/bin/*

ARG     DEP_TAG=v0.3.2
RUN     go get -d github.com/golang/dep/cmd/dep && \
        cd /go/src/github.com/golang/dep && \
        git checkout -q "$DEP_TAG" && \
        go build -v -o /usr/bin/dep ./cmd/dep && \
        rm -rf /go/src/* /go/pkg/* /go/bin/*

WORKDIR /go/src/github.com/gotestyourself/gotestyourself
ENV     CGO_ENABLED=0

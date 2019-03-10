
ARG     GOLANG_VERSION
FROM    golang:${GOLANG_VERSION:-1.12-alpine} as golang
RUN     apk add -U curl git bash
WORKDIR /go/src/gotest.tools
ENV     CGO_ENABLED=0 \
        PS1="# " \
        GO111MODULE=on

FROM    golang as tools
RUN     go get github.com/dnephin/filewatcher@v0.3.2

ARG     DEP_TAG=v0.4.1
RUN     export GO111MODULE=off; \
        go get -d github.com/golang/dep/cmd/dep && \
        cd /go/src/github.com/golang/dep && \
        git checkout -q "$DEP_TAG" && \
        go build -o /usr/bin/dep ./cmd/dep

RUN     go get gotest.tools/gotestsum@v0.3.3
RUN     wget -O- -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s && \
            mv bin/golangci-lint /go/bin


FROM    golang as dev
COPY    --from=tools /go/bin/filewatcher /usr/bin/filewatcher
COPY    --from=tools /usr/bin/dep /usr/bin/dep
COPY    --from=tools /go/bin/gotestsum /usr/bin/gotestsum
COPY    --from=tools /go/bin/golangci-lint /usr/bin/golangci-lint


FROM    dev as dev-with-source
COPY    . .

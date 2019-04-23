# -- multistage docker build: stage #1: build stage
FROM golang:1.12-alpine AS build

RUN mkdir -p /go/src/github.com/daglabs/btcd

WORKDIR /go/src/github.com/daglabs/btcd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev
RUN go get -u golang.org/x/lint/golint \
      github.com/kisielk/errcheck \
      github.com/opennota/check/cmd/aligncheck \
      github.com/opennota/check/cmd/structcheck \
      github.com/opennota/check/cmd/varcheck

# GO111MODULE=on forces Go to use the go-module system
# TODO: remove this once Go 1.13 is released
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN TEST_DIRS=`go list -f {{.Dir}} ./... | grep -v /vendor/` GOFMT_RESULT=`gofmt -l $TEST_DIRS`; echo $GOFMT_RESULT; test -z "$GOFMT_RESULT"
RUN go vet ./...
RUN TEST_DIRS=`go list -f {{.Dir}} ./... | grep -v /vendor/` golint -set_exit_status $TEST_DIRS
# RUN aligncheck ./...
# RUN structcheck -e ./...
# RUN varcheck -e ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o btcd .

# Remove the line below and uncomment the line after it for testing with coverage
RUN go test -gcflags='-l' ./...
# RUN ./test.sh

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/daglabs/btcd/btcd /app/
COPY --from=build /go/src/github.com/daglabs/btcd/docker/rpc.cert /app/
COPY --from=build /go/src/github.com/daglabs/btcd/docker/rpc.key /app/
COPY --from=build /go/src/github.com/daglabs/btcd/sample-btcd.conf /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

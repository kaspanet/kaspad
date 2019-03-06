# -- multistage docker build: stage #1: build stage
FROM golang:1.12-alpine AS build

RUN mkdir -p /go/src/github.com/daglabs/btcd

WORKDIR /go/src/github.com/daglabs/btcd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN go get -u golang.org/x/lint/golint \
      github.com/kisielk/errcheck \
      github.com/opennota/check/cmd/aligncheck \
      github.com/opennota/check/cmd/structcheck \
      github.com/opennota/check/cmd/varcheck

COPY ./Gopkg.* ./

RUN dep ensure -v -vendor-only

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
COPY --from=build /go/src/github.com/daglabs/btcd/version.txt /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

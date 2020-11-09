# -- multistage docker build: stage #1: build stage
FROM golang:1.14-alpine AS build

RUN mkdir -p /go/src/github.com/kaspanet/kaspad

WORKDIR /go/src/github.com/kaspanet/kaspad

RUN apk add --no-cache curl git openssh binutils gcc musl-dev
RUN go get -u golang.org/x/lint/golint \
              honnef.co/go/tools/cmd/staticcheck 

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

WORKDIR /go/src/github.com/kaspanet/kaspad/cmd/kaspactl

RUN GOFMT_RESULT=`go fmt ./...`; echo $GOFMT_RESULT; test -z "$GOFMT_RESULT"
RUN go vet ./...
RUN golint -set_exit_status ./...
RUN staticcheck -checks SA4006 ./...
RUN GOOS=linux go build -a -installsuffix cgo -o kaspactl .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/kaspanet/kaspad/cmd/kaspactl/kaspactl /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

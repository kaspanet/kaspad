FROM golang:alpine

RUN mkdir -p /go/src/github.com/daglabs/btcd

WORKDIR /go/src/github.com/daglabs/btcd

RUN apk add --no-cache curl git openssh
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY ./Gopkg.* ./

RUN dep ensure -v --vendor-only

COPY . .

RUN go install -v ./...

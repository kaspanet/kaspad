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

COPY . .

# GO111MODULE=on forces Go to use the go-module system
# TODO: remove this once Go 1.13 is released
ENV GO111MODULE=on

RUN go vet ./...
# RUN aligncheck ./...
# RUN structcheck -e ./...
# RUN varcheck -e ./...
RUN cd dnsseeder && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dnsseeder .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/daglabs/btcd/dnsseeder/ /app/

USER nobody
CMD [ "/app/dnsseeder", "-H", "dnsseed.my.home", "-n", "192.168.16.104", "-l", "0.0.0.0:53", "-s", "localhost", "--testnet" ]
ENTRYPOINT [ "/sbin/tini", "--" ]

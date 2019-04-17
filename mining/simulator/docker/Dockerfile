# -- multistage docker build: stage #1: build stage
FROM golang:1.12-alpine AS build

RUN mkdir -p /go/src/github.com/daglabs/btcd

WORKDIR /go/src/github.com/daglabs/btcd

RUN apk add --no-cache curl git

COPY . .

# GO111MODULE=on forces Go to use the go-module system
# TODO: remove this once Go 1.13 is released
ENV GO111MODULE=on

RUN cd mining/simulator && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mining_simulator .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache tini

COPY --from=build /go/src/github.com/daglabs/btcd/mining/simulator/mining_simulator /app/

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["/app/mining_simulator"]

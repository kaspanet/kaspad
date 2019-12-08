# -- multistage docker build: stage #1: build stage
FROM golang:1.13-alpine AS build

RUN mkdir -p /go/src/github.com/kaspanet/kaspad

WORKDIR /go/src/github.com/kaspanet/kaspad

RUN apk add --no-cache curl git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd dnsseeder && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dnsseeder .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache tini

COPY --from=build /go/src/github.com/kaspanet/kaspad/dnsseeder/ /app/

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/dnsseeder"]

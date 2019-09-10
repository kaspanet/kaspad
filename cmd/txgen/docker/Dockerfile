# -- multistage docker build: stage #1: build stage
FROM golang:1.13-alpine AS build

RUN mkdir -p /go/src/github.com/daglabs/btcd

WORKDIR /go/src/github.com/daglabs/btcd

RUN apk add --no-cache curl git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd cmd/txgen && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o txgen .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache tini

COPY --from=build /go/src/github.com/daglabs/btcd/cmd/txgen/txgen /app/

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["/app/txgen"]

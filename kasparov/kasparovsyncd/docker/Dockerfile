# -- multistage docker build: stage #1: build stage
FROM golang:1.13-alpine AS build

RUN mkdir -p /go/src/github.com/kaspanet/kaspad

WORKDIR /go/src/github.com/kaspanet/kaspad

RUN apk add --no-cache curl git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd kasparov/syncd && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kasparov-syncd .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache tini

COPY --from=build /go/src/github.com/kaspanet/kaspad/kasparov/kasparovsync/ /app/
COPY --from=build /go/src/github.com/kaspanet/kaspad/kasparov/database/migrations/ /database/migrations/

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/kasparov-syncd"]

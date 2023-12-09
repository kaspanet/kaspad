# -- multistage docker build: stage #1: build stage
FROM golang:1.19-alpine AS build

RUN mkdir -p /go/src/github.com/zoomy-network/zoomyd

WORKDIR /go/src/github.com/zoomy-network/zoomyd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

<<<<<<< Updated upstream:cmd/kaspactl/docker/Dockerfile
WORKDIR /go/src/github.com/zoomy-network/zoomyd/cmd/zoomyctl
=======
WORKDIR /go/src/github.com/zoomy-network/zoomyd/cmd/kaspactl
>>>>>>> Stashed changes:cmd/zoomyctl/docker/Dockerfile

RUN GOOS=linux go build -a -installsuffix cgo -o zoomyctl .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

<<<<<<< Updated upstream:cmd/kaspactl/docker/Dockerfile
COPY --from=build /go/src/github.com/zoomy-network/zoomyd/cmd/zoomyctl/zoomyctl /app/
=======
COPY --from=build /go/src/github.com/zoomy-network/zoomyd/cmd/kaspactl/kaspactl /app/
>>>>>>> Stashed changes:cmd/zoomyctl/docker/Dockerfile

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

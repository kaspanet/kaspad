# -- multistage docker build: stage #1: build stage
FROM golang:1.18-alpine AS build

RUN mkdir -p /go/src/github.com/zoomy-network/zoomyd

WORKDIR /go/src/github.com/zoomy-network/zoomyd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

<<<<<<< Updated upstream:cmd/kaspawallet/docker/Dockerfile
WORKDIR /go/src/github.com/zoomy-network/zoomyd/cmd/zoomywallet
RUN GOOS=linux go build -a -installsuffix cgo -o zoomywallet .
=======
WORKDIR /go/src/github.com/zoomy-network/zoomyd/cmd/kaspawallet
RUN GOOS=linux go build -a -installsuffix cgo -o kaspawallet .
>>>>>>> Stashed changes:cmd/zoomywallet/docker/Dockerfile

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

<<<<<<< Updated upstream:cmd/kaspawallet/docker/Dockerfile
COPY --from=build /go/src/github.com/zoomy-network/zoomyd/cmd/zoomywallet/zoomywallet /app/
=======
COPY --from=build /go/src/github.com/zoomy-network/zoomyd/cmd/kaspawallet/kaspawallet /app/
>>>>>>> Stashed changes:cmd/zoomywallet/docker/Dockerfile

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

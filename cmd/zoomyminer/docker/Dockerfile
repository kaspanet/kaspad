# -- multistage docker build: stage #1: build stage
FROM golang:1.19-alpine AS build

RUN mkdir -p /go/src/github.com/zoomy-network/zoomyd

WORKDIR /go/src/github.com/zoomy-network/zoomyd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

<<<<<<< Updated upstream:cmd/kaspaminer/docker/Dockerfile
WORKDIR /go/src/github.com/zoomy-network/zoomyd/cmd/zoomyminer
RUN GOOS=linux go build -a -installsuffix cgo -o zoomyminer .
=======
WORKDIR /go/src/github.com/zoomy-network/zoomyd/cmd/kaspaminer
RUN GOOS=linux go build -a -installsuffix cgo -o kaspaminer .
>>>>>>> Stashed changes:cmd/zoomyminer/docker/Dockerfile

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

<<<<<<< Updated upstream:cmd/kaspaminer/docker/Dockerfile
COPY --from=build /go/src/github.com/zoomy-network/zoomyd/cmd/zoomyminer/zoomyminer /app/
=======
COPY --from=build /go/src/github.com/zoomy-network/zoomyd/cmd/kaspaminer/kaspaminer /app/
>>>>>>> Stashed changes:cmd/zoomyminer/docker/Dockerfile

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

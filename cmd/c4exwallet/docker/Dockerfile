# -- multistage docker build: stage #1: build stage
FROM golang:1.18-alpine AS build

RUN mkdir -p /go/src/github.com/c4ei/c4exd

WORKDIR /go/src/github.com/c4ei/c4exd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

WORKDIR /go/src/github.com/c4ei/yunseokyeol/cmd/c4exwallet
RUN GOOS=linux go build -a -installsuffix cgo -o c4exwallet .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/c4ei/yunseokyeol/cmd/c4exwallet/c4exwallet /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

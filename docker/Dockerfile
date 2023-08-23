# -- multistage docker build: stage #1: build stage
FROM golang:1.19-alpine AS build

RUN mkdir -p /go/src/github.com/c4ei/c4exd

WORKDIR /go/src/github.com/c4ei/c4exd

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .


# Cache c4exd dependencies
RUN go mod download

COPY . .

RUN go build $FLAGS -o c4exd .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/c4ei/yunseokyeol/c4exd /app/
COPY --from=build /go/src/github.com/c4ei/yunseokyeol/infrastructure/config/sample-c4exd.conf /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

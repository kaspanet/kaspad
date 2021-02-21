# -- multistage docker build: stage #1: build stage
FROM golang:1.16-alpine AS build

RUN mkdir -p /go/src/github.com/kaspanet/kaspad

WORKDIR /go/src/github.com/kaspanet/kaspad

RUN apk add --no-cache curl git openssh binutils gcc musl-dev
RUN go get -u golang.org/x/lint/golint \
      github.com/kisielk/errcheck \
      github.com/opennota/check/cmd/aligncheck \
      github.com/opennota/check/cmd/structcheck \
      github.com/opennota/check/cmd/varcheck \
      honnef.co/go/tools/cmd/staticcheck 

COPY go.mod .
COPY go.sum .

COPY . .

RUN ./build_and_test.sh

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/kaspanet/kaspad/kaspad /app/
COPY --from=build /go/src/github.com/kaspanet/kaspad/infrastructure/config/sample-kaspad.conf /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]

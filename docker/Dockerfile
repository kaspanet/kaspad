# -- multistage docker build: stage #1: build stage
FROM golang:1.19-alpine AS build

RUN mkdir -p /go/src/github.com/fabbez/topiad

WORKDIR /go/src/github.com/fabbez/topiad

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .


# Cache topiad dependencies
RUN go mod download

COPY . .

RUN go build $FLAGS -o topiad .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/fabbez/topiad/topiad /app/
COPY --from=build /go/src/github.com/fabbez/topiad/infrastructure/config/sample-topiad.conf /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]
